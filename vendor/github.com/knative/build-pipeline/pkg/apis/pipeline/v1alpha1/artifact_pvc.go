/*
Copyright 2018 The Knative Authors.

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
	pvcDir        = "/pvc"
	bashNoopImage = flag.String("bash-noop-image", "override-with-bash-noop:latest", "The container image containing bash shell")
)

// ArtifactPVC represents the pvc created by the pipelinerun
// for artifacts temporary storage
type ArtifactPVC struct {
	Name string
}

// GetType returns the type of the artifact storage
func (p *ArtifactPVC) GetType() string {
	return ArtifactStoragePVCType
}

// StorageBasePath returns the path to be used to store artifacts in a pipelinerun temporary storage
func (p *ArtifactPVC) StorageBasePath(pr *PipelineRun) string {
	return pvcDir
}

// GetCopyFromContainerSpec returns a container used to download artifacts from temporary storage
func (p *ArtifactPVC) GetCopyFromContainerSpec(name, sourcePath, destinationPath string) []corev1.Container {
	return []corev1.Container{{
		Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("source-copy-%s", name)),
		Image: *bashNoopImage,
		Args:  []string{"-args", strings.Join([]string{"cp", "-r", fmt.Sprintf("%s/.", sourcePath), destinationPath}, " ")},
	}}
}

// GetCopyToContainerSpec returns a container used to upload artifacts for temporary storage
func (p *ArtifactPVC) GetCopyToContainerSpec(name, sourcePath, destinationPath string) []corev1.Container {
	return []corev1.Container{{
		Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("source-mkdir-%s", name)),
		Image: *bashNoopImage,
		Args: []string{
			"-args", strings.Join([]string{"mkdir", "-p", destinationPath}, " "),
		},
		VolumeMounts: []corev1.VolumeMount{getPvcMount(p.Name)},
	}, {
		Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("source-copy-%s", name)),
		Image: *bashNoopImage,
		Args: []string{
			"-args", strings.Join([]string{"cp", "-r", fmt.Sprintf("%s/.", sourcePath), destinationPath}, " "),
		},
		VolumeMounts: []corev1.VolumeMount{getPvcMount(p.Name)},
	}}
}

func getPvcMount(name string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,   // taskrun pvc name
		MountPath: pvcDir, // nothing should be mounted here
	}
}

// CreateDirContainer returns a container step to create a dir
func CreateDirContainer(name, destinationPath string) corev1.Container {
	return corev1.Container{
		Name:  names.SimpleNameGenerator.GenerateName(fmt.Sprintf("create-dir-%s", name)),
		Image: *bashNoopImage,
		Args:  []string{"-args", strings.Join([]string{"mkdir", "-p", destinationPath}, " ")},
	}
}

// GetSecretsVolumes returns the list of volumes for secrets to be mounted
// on pod
func (p *ArtifactPVC) GetSecretsVolumes() []corev1.Volume {
	return nil
}
