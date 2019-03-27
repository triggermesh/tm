/*
Copyright 2018 The Knative Authors
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

package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	fakeclientset "github.com/knative/build-pipeline/pkg/client/clientset/versioned/fake"
	informers "github.com/knative/build-pipeline/pkg/client/informers/externalversions"
	listers "github.com/knative/build-pipeline/pkg/client/listers/pipeline/v1alpha1"
	"github.com/knative/build-pipeline/pkg/logging"
	"github.com/knative/build-pipeline/test/names"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

var (
	pipelineResourceLister listers.PipelineResourceLister
	logger                 *zap.SugaredLogger
)

func setUp() {
	logger, _ = logging.NewLogger("", "")
	fakeClient := fakeclientset.NewSimpleClientset()
	sharedInfomer := informers.NewSharedInformerFactory(fakeClient, 0)
	pipelineResourceInformer := sharedInfomer.Tekton().V1alpha1().PipelineResources()
	pipelineResourceLister = pipelineResourceInformer.Lister()

	rs := []*v1alpha1.PipelineResource{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "the-git",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
			Params: []v1alpha1.Param{{
				Name:  "Url",
				Value: "https://github.com/grafeas/kritis",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "the-git-with-branch",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
			Params: []v1alpha1.Param{{
				Name:  "Url",
				Value: "https://github.com/grafeas/kritis",
			}, {
				Name:  "Revision",
				Value: "branch",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster2",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "cluster",
			Params: []v1alpha1.Param{{
				Name:  "Name",
				Value: "cluster2",
			}, {
				Name:  "Url",
				Value: "http://10.10.10.10",
			}},
			SecretParams: []v1alpha1.SecretParam{{
				FieldName:  "cadata",
				SecretKey:  "cadatakey",
				SecretName: "secret1",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster3",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "cluster",
			Params: []v1alpha1.Param{{
				Name:  "name",
				Value: "cluster3",
			}, {
				Name:  "Url",
				Value: "http://10.10.10.10",
			}, {
				Name: "CAdata",
				// echo "my-ca-cert" | base64
				Value: "bXktY2EtY2VydAo=",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "storage1",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "storage",
			Params: []v1alpha1.Param{{
				Name:  "Location",
				Value: "gs://fake-bucket/rules.zip",
			}, {
				Name:  "Type",
				Value: "gcs",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "storage-gcs-keys",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "storage",
			Params: []v1alpha1.Param{{
				Name:  "Location",
				Value: "gs://fake-bucket/rules.zip",
			}, {
				Name:  "Type",
				Value: "gcs",
			}, {
				Name:  "Dir",
				Value: "true",
			}},
			SecretParams: []v1alpha1.SecretParam{{
				SecretKey:  "key.json",
				SecretName: "secret-name",
				FieldName:  "GOOGLE_APPLICATION_CREDENTIALS",
			}, {
				SecretKey:  "token",
				SecretName: "secret-name2",
				FieldName:  "GOOGLE_TOKEN",
			}},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "storage-gcs-invalid",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "storage",
			Params: []v1alpha1.Param{{
				Name:  "Location",
				Value: "gs://fake-bucket/rules",
			}, {
				Name:  "Type",
				Value: "non-existent",
			}},
		},
	}}
	for _, r := range rs {
		pipelineResourceInformer.Informer().GetIndexer().Add(r)
	}
}

func build() *buildv1alpha1.Build {
	boolTrue := true
	return &buildv1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Build",
			APIVersion: "build.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-from-repo",
			Namespace: "marshmallow",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "tekton.dev/v1alpha1",
				Kind:               "TaskRun",
				Name:               "build-from-repo-run",
				Controller:         &boolTrue,
				BlockOwnerDeletion: &boolTrue,
			}},
		},
		Spec: buildv1alpha1.BuildSpec{},
	}
}
func TestAddResourceToBuild(t *testing.T) {
	task := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-from-repo",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "gitspace",
					Type: "git",
				}},
			},
		},
	}
	taskWithMultipleGitSources := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-from-repo",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "gitspace",
					Type: "git",
				}, {
					Name: "git-duplicate-space",
					Type: "git",
				}},
			},
		},
	}
	taskWithTargetPath := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "task-with-targetpath",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name:       "workspace",
					Type:       "gcs",
					TargetPath: "gcs-dir",
				}},
			},
		},
	}

	taskRun := &v1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-from-repo-run",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskRunSpec{
			TaskRef: &v1alpha1.TaskRef{
				Name: "simpleTask",
			},
			Inputs: v1alpha1.TaskRunInputs{
				Resources: []v1alpha1.TaskResourceBinding{{
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "the-git",
					},
					Name: "gitspace",
				}},
			},
		},
	}

	for _, c := range []struct {
		desc    string
		task    *v1alpha1.Task
		taskRun *v1alpha1.TaskRun
		build   *buildv1alpha1.Build
		wantErr bool
		want    buildv1alpha1.BuildSpec
	}{{
		desc:    "simple with default revision",
		task:    task,
		taskRun: taskRun,
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:       "git-source-the-git-9l9zj",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "master", "-path", "/workspace/gitspace"},
				WorkingDir: "/workspace",
			}},
		},
	}, {
		desc: "simple with branch",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "simpleTask",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git-with-branch",
						},
						Name: "gitspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:       "git-source-the-git-with-branch-9l9zj",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "branch", "-path", "/workspace/gitspace"},
				WorkingDir: "/workspace",
			}},
		},
	}, {
		desc: "same git input resource for task with diff resource name",
		task: taskWithMultipleGitSources,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "simpleTask",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git-with-branch",
						},
						Name: "gitspace",
					}, {
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git-with-branch",
						},
						Name: "git-duplicate-space",
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:       "git-source-the-git-with-branch-mz4c7",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "branch", "-path", "/workspace/git-duplicate-space"},
				WorkingDir: "/workspace",
			}, {
				Name:       "git-source-the-git-with-branch-9l9zj",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "branch", "-path", "/workspace/gitspace"},
				WorkingDir: "/workspace",
			}},
		},
	}, {
		desc: "set revision to default value 1",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "simpleTask",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git",
						},
						Name: "gitspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:       "git-source-the-git-9l9zj",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "master", "-path", "/workspace/gitspace"},
				WorkingDir: "/workspace",
			}},
		},
	}, {
		desc: "set revision to provdided branch",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "simpleTask",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git-with-branch",
						},
						Name: "gitspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:       "git-source-the-git-with-branch-9l9zj",
				Image:      "override-with-git:latest",
				Args:       []string{"-url", "https://github.com/grafeas/kritis", "-revision", "branch", "-path", "/workspace/gitspace"},
				WorkingDir: "/workspace",
			}},
		},
	}, {
		desc: "git resource as input from previous task",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-git",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					Kind: "PipelineRun",
					Name: "pipelinerun",
				}},
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git",
						},
						Name:  "gitspace",
						Paths: []string{"prev-task-path"},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "create-dir-gitspace-mz4c7",
				Image: "override-with-bash-noop:latest",
				Args:  []string{"-args", "mkdir -p /workspace/gitspace"},
			}, {
				Name:         "source-copy-gitspace-0-9l9zj",
				Image:        "override-with-bash-noop:latest",
				Args:         []string{"-args", "cp -r prev-task-path/. /workspace/gitspace"},
				VolumeMounts: []corev1.VolumeMount{{MountPath: "/pvc", Name: "pipelinerun-pvc"}},
			}},
			Volumes: []corev1.Volume{{
				Name: "pipelinerun-pvc",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "pipelinerun-pvc"},
				},
			}},
		},
	}, {
		desc: "storage resource as input with target path",
		task: taskWithTargetPath,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-gcs",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage1",
						},
						Name: "workspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "create-dir-storage1-9l9zj",
				Image: "override-with-bash-noop:latest",
				Args:  []string{"-args", "mkdir -p /workspace/gcs-dir"},
			}, {
				Name:  "fetch-storage1-mz4c7",
				Image: "override-with-gsutil-image:latest",
				Args:  []string{"-args", "cp gs://fake-bucket/rules.zip /workspace/gcs-dir"},
			}},
		},
	}, {
		desc: "storage resource as input from previous task",
		task: taskWithTargetPath,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-gcs",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					Kind: "PipelineRun",
					Name: "pipelinerun",
				}},
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage1",
						},
						Name:  "workspace",
						Paths: []string{"prev-task-path"},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "create-dir-workspace-mz4c7",
				Image: "override-with-bash-noop:latest",
				Args:  []string{"-args", "mkdir -p /workspace/gcs-dir"},
			}, {
				Name:         "source-copy-workspace-0-9l9zj",
				Image:        "override-with-bash-noop:latest",
				Args:         []string{"-args", "cp -r prev-task-path/. /workspace/gcs-dir"},
				VolumeMounts: []corev1.VolumeMount{{MountPath: "/pvc", Name: "pipelinerun-pvc"}},
			}},
			Volumes: []corev1.Volume{{
				Name: "pipelinerun-pvc",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "pipelinerun-pvc"},
				},
			}},
		},
	}, {
		desc: "invalid gcs resource type name",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-invalid-gcs",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage-gcs-invalid",
						},
						Name: "workspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: true,
	}, {
		desc: "invalid gcs resource type name",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-invalid-gcs",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage-gcs-invalid",
						},
						Name: "workspace",
					}},
				},
			},
		},
		build:   build(),
		wantErr: true,
	}, {
		desc: "invalid resource name",
		task: &v1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "workspace-invalid",
						Type: "git",
					}},
				},
			},
		},
		taskRun: taskRun,
		build:   build(),
		wantErr: true,
	}, {
		desc: "cluster resource with plain text",
		task: &v1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "target-cluster",
						Type: "cluster",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "build-from-repo",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "target-cluster",
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "cluster3",
						},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "kubeconfig-9l9zj",
				Image: "override-with-kubeconfig-writer:latest",
				Args: []string{
					"-clusterConfig", `{"name":"cluster3","type":"cluster","url":"http://10.10.10.10","revision":"","username":"","password":"","token":"","Insecure":false,"cadata":"bXktY2EtY2VydAo=","secrets":null}`,
				},
			}},
		},
	}, {
		desc: "cluster resource with secrets",
		task: &v1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "target-cluster",
						Type: "cluster",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				TaskRef: &v1alpha1.TaskRef{
					Name: "build-from-repo",
				},
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "target-cluster",
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "cluster2",
						},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "kubeconfig-9l9zj",
				Image: "override-with-kubeconfig-writer:latest",
				Args: []string{
					"-clusterConfig", `{"name":"cluster2","type":"cluster","url":"http://10.10.10.10","revision":"","username":"","password":"","token":"","Insecure":false,"cadata":null,"secrets":[{"fieldName":"cadata","secretKey":"cadatakey","secretName":"secret1"}]}`,
				},
				Env: []corev1.EnvVar{{
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secret1",
							},
							Key: "cadatakey",
						},
					},
					Name: "CADATA",
				}},
			}},
		},
	}} {
		t.Run(c.desc, func(t *testing.T) {
			setUp()
			names.TestingSeed()
			fakekubeclient := fakek8s.NewSimpleClientset()
			got, err := AddInputResource(fakekubeclient, c.build, c.task.Name, &c.task.Spec, c.taskRun, pipelineResourceLister, logger)
			if (err != nil) != c.wantErr {
				t.Errorf("Test: %q; AddInputResource() error = %v, WantErr %v", c.desc, err, c.wantErr)
			}
			if got != nil {
				if d := cmp.Diff(got.Spec, c.want); d != "" {
					t.Errorf("Diff:\n%s", d)
				}
			}
		})
	}
}

func Test_StorageInputResource(t *testing.T) {
	boolTrue := true
	for _, c := range []struct {
		desc    string
		task    *v1alpha1.Task
		taskRun *v1alpha1.TaskRun
		build   *buildv1alpha1.Build
		wantErr bool
		want    *buildv1alpha1.Build
	}{{
		desc: "inputs with no resource spec and resource ref",
		task: &v1alpha1.Task{
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "gcs-input-resource",
						Type: "storage",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "gcs-input-resource",
					}},
				},
			},
		},
		build:   nil,
		wantErr: true,
	}, {
		desc: "inputs with both resource spec and resource ref",
		task: &v1alpha1.Task{
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "gcs-input-resource",
						Type: "storage",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "gcs-input-resource",
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage-gcs-keys",
						},
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: v1alpha1.PipelineResourceTypeStorage,
						},
					}},
				},
			},
		},
		build:   nil,
		wantErr: true,
	}, {
		desc: "inputs with resource spec and no resource ref",
		task: &v1alpha1.Task{
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "gcs-input-resource",
						Type: "storage",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "gcs-input-resource",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: v1alpha1.PipelineResourceTypeStorage,
							Params: []v1alpha1.Param{{
								Name:  "Location",
								Value: "gs://fake-bucket/rules.zip",
							}, {
								Name:  "Type",
								Value: "gcs",
							}},
						},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: &buildv1alpha1.Build{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Build",
				APIVersion: "build.knative.dev/v1alpha1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         "tekton.dev/v1alpha1",
					Kind:               "TaskRun",
					Name:               "build-from-repo-run",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: buildv1alpha1.BuildSpec{
				Steps: []corev1.Container{{
					Name:  "create-dir-gcs-input-resource-9l9zj",
					Image: "override-with-bash-noop:latest",
					Args:  []string{"-args", "mkdir -p /workspace/gcs-input-resource"},
				}, {
					Name:  "fetch-gcs-input-resource-mz4c7",
					Image: "override-with-gsutil-image:latest",
					Args:  []string{"-args", "cp gs://fake-bucket/rules.zip /workspace/gcs-input-resource"},
				}},
			},
		},
	}, {
		desc: "no inputs",
		task: &v1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage",
				Namespace: "marshmallow",
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage-run",
				Namespace: "marshmallow",
			},
		},
		build:   nil,
		wantErr: false,
	}, {
		desc: "storage resource as input",
		task: &v1alpha1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskSpec{
				Inputs: &v1alpha1.Inputs{
					Resources: []v1alpha1.TaskResource{{
						Name: "gcs-input-resource",
						Type: "storage",
					}},
				},
			},
		},
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-storage-run",
				Namespace: "marshmallow",
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						Name: "gcs-input-resource",
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage-gcs-keys",
						},
					}},
				},
			},
		},
		build:   build(),
		wantErr: false,
		want: &buildv1alpha1.Build{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Build",
				APIVersion: "build.knative.dev/v1alpha1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "build-from-repo",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         "tekton.dev/v1alpha1",
					Kind:               "TaskRun",
					Name:               "build-from-repo-run",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: buildv1alpha1.BuildSpec{
				Steps: []corev1.Container{{
					Name:  "create-dir-storage-gcs-keys-9l9zj",
					Image: "override-with-bash-noop:latest",
					Args:  []string{"-args", "mkdir -p /workspace/gcs-input-resource"},
				}, {
					Name:  "fetch-storage-gcs-keys-mz4c7",
					Image: "override-with-gsutil-image:latest",
					Args:  []string{"-args", "cp -r gs://fake-bucket/rules.zip/** /workspace/gcs-input-resource"},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "volume-storage-gcs-keys-secret-name", MountPath: "/var/secret/secret-name"},
					},
					Env: []corev1.EnvVar{
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/secret-name/key.json"},
					},
				}},
				Volumes: []corev1.Volume{{
					Name:         "volume-storage-gcs-keys-secret-name",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "secret-name"}},
				}, {
					Name:         "volume-storage-gcs-keys-secret-name2",
					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "secret-name2"}},
				}},
			},
		},
	}} {
		t.Run(c.desc, func(t *testing.T) {
			names.TestingSeed()
			setUp()
			fakekubeclient := fakek8s.NewSimpleClientset()
			got, err := AddInputResource(fakekubeclient, c.build, c.task.Name, &c.task.Spec, c.taskRun, pipelineResourceLister, logger)
			if (err != nil) != c.wantErr {
				t.Errorf("Test: %q; AddInputResource() error = %v, WantErr %v", c.desc, err, c.wantErr)
			}
			if d := cmp.Diff(got, c.want); d != "" {
				t.Errorf("Diff:\n%s", d)
			}
		})
	}
}

func TestAddStepsToBuild_WithBucketFromConfigMap(t *testing.T) {
	task := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-from-repo",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "gitspace",
					Type: "git",
				}},
			},
		},
	}
	taskWithTargetPath := &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "task-with-targetpath",
			Namespace: "marshmallow",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name:       "workspace",
					Type:       "gcs",
					TargetPath: "gcs-dir",
				}},
			},
		},
	}

	for _, c := range []struct {
		desc    string
		task    *v1alpha1.Task
		taskRun *v1alpha1.TaskRun
		build   *buildv1alpha1.Build
		want    buildv1alpha1.BuildSpec
	}{{
		desc: "git resource as input from previous task - copy to bucket",
		task: task,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-git",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					Kind: "PipelineRun",
					Name: "pipelinerun",
				}},
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "the-git",
						},
						Name:  "gitspace",
						Paths: []string{"prev-task-path"},
					}},
				},
			},
		},
		build: build(),
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "artifact-dest-mkdir-gitspace-0-mssqb",
				Image: "override-with-bash-noop:latest",
				Args:  []string{"-args", "mkdir -p /workspace/gitspace"},
			}, {
				Name:  "artifact-copy-from-gitspace-0-78c5n",
				Image: "override-with-gsutil-image:latest",
				Args:  []string{"-args", "cp -r gs://fake-bucket/prev-task-path/** /workspace/gitspace"},
			}},
		},
	}, {
		desc: "storage resource as input from previous task - copy from bucket",
		task: taskWithTargetPath,
		taskRun: &v1alpha1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "get-from-gcs",
				Namespace: "marshmallow",
				OwnerReferences: []metav1.OwnerReference{{
					Kind: "PipelineRun",
					Name: "pipelinerun",
				}},
			},
			Spec: v1alpha1.TaskRunSpec{
				Inputs: v1alpha1.TaskRunInputs{
					Resources: []v1alpha1.TaskResourceBinding{{
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name: "storage1",
						},
						Name:  "workspace",
						Paths: []string{"prev-task-path"},
					}},
				},
			},
		},
		build: build(),
		want: buildv1alpha1.BuildSpec{
			Steps: []corev1.Container{{
				Name:  "artifact-dest-mkdir-workspace-0-6nl7g",
				Image: "override-with-bash-noop:latest",
				Args:  []string{"-args", "mkdir -p /workspace/gcs-dir"},
			}, {
				Name:  "artifact-copy-from-workspace-0-j2tds",
				Image: "override-with-gsutil-image:latest",
				Args:  []string{"-args", "cp -r gs://fake-bucket/prev-task-path/** /workspace/gcs-dir"},
			}},
		},
	}} {
		t.Run(c.desc, func(t *testing.T) {
			setUp()
			fakekubeclient := fakek8s.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "tekton-pipelines",
						Name:      v1alpha1.BucketConfigName,
					},
					Data: map[string]string{
						v1alpha1.BucketLocationKey: "gs://fake-bucket",
					},
				},
			)
			got, err := AddInputResource(fakekubeclient, c.build, c.task.Name, &c.task.Spec, c.taskRun, pipelineResourceLister, logger)
			if err != nil {
				t.Errorf("Test: %q; AddInputResource() error = %v", c.desc, err)
			}
			if d := cmp.Diff(got.Spec, c.want); d != "" {
				t.Errorf("Diff:\n%s", d)
			}
		})
	}
}
