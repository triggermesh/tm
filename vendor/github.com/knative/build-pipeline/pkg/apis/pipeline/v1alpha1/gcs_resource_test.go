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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knative/build-pipeline/test/names"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Invalid_NewStorageResource(t *testing.T) {
	testcases := []struct {
		name             string
		pipelineResource *PipelineResource
	}{{
		name: "wrong-resource-type",
		pipelineResource: &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gcs-resource",
			},
			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeGit,
			},
		},
	}, {
		name: "unimplemented type",
		pipelineResource: &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gcs-resource",
			},
			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeStorage,
				Params: []Param{{
					Name:  "Location",
					Value: "gs://fake-bucket",
				}, {
					Name:  "type",
					Value: "non-existent-type",
				}},
			},
		},
	}, {
		name: "no type",
		pipelineResource: &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gcs-resource",
			},
			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeStorage,
				Params: []Param{{
					Name:  "Location",
					Value: "gs://fake-bucket",
				}},
			},
		},
	}, {
		name: "no location params",
		pipelineResource: &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gcs-resource-with-no-location-param",
			},
			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeStorage,
				Params: []Param{{
					Name:  "NotLocation",
					Value: "doesntmatter",
				}, {
					Name:  "type",
					Value: "gcs",
				}},
			},
		},
	}, {
		name: "location param with empty value",
		pipelineResource: &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gcs-resource-with-empty-location-param",
			},
			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeStorage,
				Params: []Param{{
					Name:  "Location",
					Value: "",
				}, {
					Name:  "type",
					Value: "gcs",
				}},
			},
		},
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewStorageResource(tc.pipelineResource)
			if err == nil {
				t.Error("Expected error creating GCS resource")
			}
		})
	}
}

func Test_Valid_NewGCSResource(t *testing.T) {
	pr := &PipelineResource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gcs-resource",
		},
		Spec: PipelineResourceSpec{
			Type: PipelineResourceTypeStorage,
			Params: []Param{{
				Name:  "Location",
				Value: "gs://fake-bucket",
			}, {
				Name:  "type",
				Value: "gcs",
			}, {
				Name:  "dir",
				Value: "anything",
			}},
			SecretParams: []SecretParam{{
				SecretKey:  "secretKey",
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
			}},
		},
	}
	expectedGCSResource := &GCSResource{
		Name:     "gcs-resource",
		Location: "gs://fake-bucket",
		Type:     PipelineResourceTypeStorage,
		TypeDir:  true,
		Secrets: []SecretParam{{
			SecretName: "secretName",
			SecretKey:  "secretKey",
			FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
		}},
	}

	gcsRes, err := NewGCSResource(pr)
	if err != nil {
		t.Fatalf("Unexpected error creating GCS resource: %s", err)
	}
	if d := cmp.Diff(expectedGCSResource, gcsRes); d != "" {
		t.Errorf("Mismatch of GCS resource: %s", d)
	}
}

func Test_GCSGetReplacements(t *testing.T) {
	gcsResource := &GCSResource{
		Name:     "gcs-resource",
		Location: "gs://fake-bucket",
		Type:     PipelineResourceTypeGCS,
	}
	expectedReplacementMap := map[string]string{
		"name":     "gcs-resource",
		"type":     "gcs",
		"location": "gs://fake-bucket",
	}
	if d := cmp.Diff(gcsResource.Replacements(), expectedReplacementMap); d != "" {
		t.Errorf("GCS Replacement map mismatch: %s", d)
	}
}

func Test_GetParams(t *testing.T) {
	pr := &PipelineResource{
		Spec: PipelineResourceSpec{
			Type: PipelineResourceTypeStorage,
			Params: []Param{{
				Name:  "type",
				Value: "gcs",
			}, {
				Name:  "location",
				Value: "gs://some-bucket.zip",
			}},
			SecretParams: []SecretParam{{
				SecretKey:  "test-secret-key",
				SecretName: "test-secret-name",
				FieldName:  "test-field-name",
			}},
		},
	}
	gcsResource, err := NewStorageResource(pr)
	if err != nil {
		t.Fatalf("Error creating storage resource: %s", err.Error())
	}
	expectedSp := []SecretParam{{
		SecretKey:  "test-secret-key",
		SecretName: "test-secret-name",
		FieldName:  "test-field-name",
	}}
	if d := cmp.Diff(gcsResource.GetSecretParams(), expectedSp); d != "" {
		t.Errorf("Error mismatch on storage secret params: %s", d)
	}
}

func Test_GetDownloadContainerSpec(t *testing.T) {
	names.TestingSeed()

	testcases := []struct {
		name           string
		gcsResource    *GCSResource
		wantContainers []corev1.Container
		wantErr        bool
	}{{
		name: "valid download protected buckets",
		gcsResource: &GCSResource{
			Name:           "gcs-valid",
			Location:       "gs://some-bucket",
			DestinationDir: "/workspace",
			TypeDir:        true,
			Secrets: []SecretParam{{
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
				SecretKey:  "key.json",
			}},
		},
		wantContainers: []corev1.Container{{
			Name:  "create-dir-gcs-valid-9l9zj",
			Image: "override-with-bash-noop:latest",
			Args:  []string{"-args", "mkdir -p /workspace"},
		}, {
			Name:  "fetch-gcs-valid-mz4c7",
			Image: "override-with-gsutil-image:latest",
			Args:  []string{"-args", "cp -r gs://some-bucket/** /workspace"},
			Env: []corev1.EnvVar{{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/secret/secretName/key.json",
			}},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "volume-gcs-valid-secretName",
				MountPath: "/var/secret/secretName",
			}},
		}},
	}, {
		name: "duplicate secret mount paths",
		gcsResource: &GCSResource{
			Name:           "gcs-valid",
			Location:       "gs://some-bucket",
			DestinationDir: "/workspace",
			Secrets: []SecretParam{{
				SecretName: "secretName",
				FieldName:  "fieldName",
				SecretKey:  "key.json",
			}, {
				SecretKey:  "key.json",
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
			}},
		},
		wantContainers: []corev1.Container{{
			Name:  "create-dir-gcs-valid-mssqb",
			Image: "override-with-bash-noop:latest",
			Args:  []string{"-args", "mkdir -p /workspace"},
		}, {
			Name:  "fetch-gcs-valid-78c5n",
			Image: "override-with-gsutil-image:latest",
			Args:  []string{"-args", "cp gs://some-bucket /workspace"},
			Env: []corev1.EnvVar{{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/secret/secretName/key.json",
			}},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "volume-gcs-valid-secretName",
				MountPath: "/var/secret/secretName",
			}},
		}},
	}, {
		name: "invalid no destination directory set",
		gcsResource: &GCSResource{
			Name:     "gcs-invalid",
			Location: "gs://some-bucket",
		},
		wantErr: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotContainers, err := tc.gcsResource.GetDownloadContainerSpec()
			if tc.wantErr && err == nil {
				t.Fatalf("Expected error to be %t but got %v:", tc.wantErr, err)
			}
			if d := cmp.Diff(gotContainers, tc.wantContainers); d != "" {
				t.Errorf("Error mismatch between download containers spec: %s", d)
			}
		})
	}
}

func Test_GetUploadContainerSpec(t *testing.T) {
	names.TestingSeed()

	testcases := []struct {
		name           string
		gcsResource    *GCSResource
		wantContainers []corev1.Container
		wantErr        bool
	}{{
		name: "valid upload to protected buckets with directory paths",
		gcsResource: &GCSResource{
			Name:           "gcs-valid",
			Location:       "gs://some-bucket",
			DestinationDir: "/workspace/",
			TypeDir:        true,
			Secrets: []SecretParam{{
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
				SecretKey:  "key.json",
			}},
		},
		wantContainers: []corev1.Container{{
			Name:  "upload-gcs-valid-9l9zj",
			Image: "override-with-gsutil-image:latest",
			Args:  []string{"-args", "cp -r /workspace/* gs://some-bucket"},
			Env:   []corev1.EnvVar{{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/secretName/key.json"}},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "volume-gcs-valid-secretName",
				MountPath: "/var/secret/secretName",
			}},
		}},
	}, {
		name: "duplicate secret mount paths",
		gcsResource: &GCSResource{
			Name:           "gcs-valid",
			Location:       "gs://some-bucket",
			DestinationDir: "/workspace",
			Secrets: []SecretParam{{
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
				SecretKey:  "key.json",
			}, {
				SecretKey:  "key.json",
				SecretName: "secretName",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
			}},
		},
		wantContainers: []corev1.Container{{
			Name:  "upload-gcs-valid-mz4c7",
			Image: "override-with-gsutil-image:latest",
			Args:  []string{"-args", "cp /workspace/* gs://some-bucket"},
			Env: []corev1.EnvVar{
				{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/secretName/key.json"},
			},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      "volume-gcs-valid-secretName",
				MountPath: "/var/secret/secretName",
			}},
		}},
	}, {
		name: "valid upload to protected buckets with single file",
		gcsResource: &GCSResource{
			Name:           "gcs-valid",
			Location:       "gs://some-bucket",
			DestinationDir: "/workspace/",
			TypeDir:        false,
		},
		wantContainers: []corev1.Container{{
			Name:  "upload-gcs-valid-mssqb",
			Image: "override-with-gsutil-image:latest",
			Args:  []string{"-args", "cp /workspace/* gs://some-bucket"},
		}},
	}, {
		name: "invalid upload with no source directory path",
		gcsResource: &GCSResource{
			Name:     "gcs-invalid",
			Location: "gs://some-bucket",
		},
		wantErr: true,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotContainers, err := tc.gcsResource.GetUploadContainerSpec()
			if tc.wantErr && err == nil {
				t.Fatalf("Expected error to be %t but got %v:", tc.wantErr, err)
			}

			if d := cmp.Diff(gotContainers, tc.wantContainers); d != "" {
				t.Errorf("Error mismatch between upload containers spec: %s", d)
			}
		})
	}
}
