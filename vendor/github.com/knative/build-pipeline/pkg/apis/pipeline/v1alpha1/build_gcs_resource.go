/*
Copyright 2019 The Knative Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"flag"
	"fmt"
	"strings"

	"github.com/knative/build-pipeline/pkg/names"

	corev1 "k8s.io/api/core/v1"
)

var (
	buildGCSFetcherImage = flag.String("build-gcs-fetcher-image", "gcr.io/cloud-builders/gcs-fetcher:latest",
		"The container image containing our GCS fetcher binary.")
	buildGCSUploaderImage = flag.String("build-gcs-uploader-image", "gcr.io/cloud-builders/gcs-uploader:latest",
		"The container image containing our GCS uploader binary.")
)

// GCSArtifactType defines a type of GCS resource.
type GCSArtifactType string

const (
	// GCSArchive indicates that resource should be fetched from a typical archive file.
	GCSArchive GCSArtifactType = "Archive"

	// GCSManifest indicates that resource should be fetched using a
	// manifest-based protocol which enables incremental source upload.
	GCSManifest GCSArtifactType = "Manifest"

	// EmptyArtifactType indicates, no artifact type is specified.
	EmptyArtifactType = ""
)

var validArtifactTypes = []string{string(GCSArchive), string(GCSManifest)}

// BuildGCSResource describes a resource in the form of an archive,
// or a source manifest describing files to fetch.
// BuildGCSResource does incremental uploads for files in  directory.

type BuildGCSResource struct {
	Name           string
	Type           PipelineResourceType
	Location       string
	DestinationDir string
	ArtifactType   GCSArtifactType
}

// NewBuildGCSResource creates a new BuildGCS resource to pass to a Task
func NewBuildGCSResource(r *PipelineResource) (*BuildGCSResource, error) {
	if r.Spec.Type != PipelineResourceTypeStorage {
		return nil, fmt.Errorf("BuildGCSResource: Cannot create a BuildGCS resource from a %s Pipeline Resource", r.Spec.Type)
	}
	if r.Spec.SecretParams != nil {
		return nil, fmt.Errorf("BuildGCSResource: %s cannot support artifacts on private bucket", r.Name)
	}
	var location string
	var aType GCSArtifactType

	for _, param := range r.Spec.Params {
		switch {
		case strings.EqualFold(param.Name, "Location"):
			location = param.Value
		case strings.EqualFold(param.Name, "ArtifactType"):
			var err error
			aType, err = getArtifactType(param.Value)
			if err != nil {
				return nil, fmt.Errorf("BuildGCSResource %s : %s", r.Name, err)
			}
		}
	}
	if location == "" {
		return nil, fmt.Errorf("BuildGCSResource: Need Location to be specified in order to create BuildGCS resource %s", r.Name)
	}
	if aType == EmptyArtifactType {
		return nil, fmt.Errorf("BuildGCSResource: Need ArtifactType to be specified in order to fetch BuildGCS resource %s", r.Name)
	}
	return &BuildGCSResource{
		Name:         r.Name,
		Type:         r.Spec.Type,
		Location:     location,
		ArtifactType: aType,
	}, nil
}

// GetName returns the name of the resource
func (s BuildGCSResource) GetName() string {
	return s.Name
}

// GetType returns the type of the resource, in this case "storage"
func (s BuildGCSResource) GetType() PipelineResourceType {
	return PipelineResourceTypeStorage
}

// GetParams get params
func (s *BuildGCSResource) GetParams() []Param { return []Param{} }

// GetSecretParams returns the resource secret params
func (s *BuildGCSResource) GetSecretParams() []SecretParam { return nil }

// Replacements is used for template replacement on an GCSResource inside of a Taskrun.
func (s *BuildGCSResource) Replacements() map[string]string {
	return map[string]string{
		"name":     s.Name,
		"type":     string(s.Type),
		"location": s.Location,
	}
}

// SetDestinationDirectory sets the destination directory at runtime like where is the resource going to be copied to
func (s *BuildGCSResource) SetDestinationDirectory(destDir string) { s.DestinationDir = destDir }

// GetDownloadContainerSpec returns an array of container specs to download gcs storage object
func (s *BuildGCSResource) GetDownloadContainerSpec() ([]corev1.Container, error) {
	if s.DestinationDir == "" {
		return nil, fmt.Errorf("BuildGCSResource: Expect Destination Directory param to be set %s", s.Name)
	}
	args := []string{"--type", string(s.ArtifactType), "--location", s.Location}
	// dest_dir is the destination directory for GCS files to be copies"
	if s.DestinationDir != "" {
		args = append(args, "--dest_dir", s.DestinationDir)
	}

	return []corev1.Container{
		CreateDirContainer(s.Name, s.DestinationDir), {
			Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("storage-fetch-%s", s.Name)),
			Image: *buildGCSFetcherImage,
			Args:  args,
		}}, nil
}

// GetUploadContainerSpec gets container spec for gcs resource to be uploaded like
// set environment variable from secret params and set volume mounts for those secrets
func (s *BuildGCSResource) GetUploadContainerSpec() ([]corev1.Container, error) {
	if s.ArtifactType != GCSManifest {
		return nil, fmt.Errorf("BuildGCSResource: Can only upload Artifacts of type Manifest: %s", s.Name)
	}
	if s.DestinationDir == "" {
		return nil, fmt.Errorf("BuildGCSResource: Expect Destination Directory param to be set %s", s.Name)
	}
	args := []string{"--location", s.Location, "--dir", s.DestinationDir}

	return []corev1.Container{{
		Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("storage-upload-%s", s.Name)),
		Image: *buildGCSUploaderImage,
		Args:  args,
	}}, nil
}

func getArtifactType(val string) (GCSArtifactType, error) {
	aType := GCSArtifactType(val)
	valid := []string{string(GCSArchive), string(GCSManifest)}
	switch aType {
	case GCSArchive:
		return aType, nil
	case GCSManifest:
		return aType, nil
	case EmptyArtifactType:
		return "", fmt.Errorf("ArtifactType is empty. Should be one of %s", strings.Join(valid, ","))
	}
	return "", fmt.Errorf("Invalid ArtifactType %s. Should be one of %s", aType, strings.Join(valid, ","))
}
