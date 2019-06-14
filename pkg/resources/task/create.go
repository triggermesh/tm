// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy creates dummy task with kaniko executor which accepts source URL and pushes resulting image to registry
func (t *Task) Deploy(clientset *client.ConfigSet) error {
	task := tekton.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
		},
		Spec: tekton.TaskSpec{
			Inputs: &tekton.Inputs{
				Resources: []tekton.TaskResource{
					{
						Name:       "sources",
						Type:       tekton.PipelineResourceType("git"),
						TargetPath: "/workspace",
					},
				},
				Params: []tekton.TaskParam{
					{
						Name:        "registry",
						Default:     "",
						Description: "Where to store resulting image",
					},
				},
			},
			Steps: t.kaniko(),
		},
	}
	return t.createOrUpdate(task, clientset)
}

func (t *Task) kaniko() []corev1.Container {
	return []corev1.Container{
		{
			Name:    "build",
			Image:   "gcr.io/kaniko-project/executor:v0.8.0",
			Command: []string{"executor"},
			Args: []string{"--dockerfile=Dockerfile",
				"--context=/workspace/workspace",
				"--destination=${inputs.params.registry}"},
		},
	}
}

func (t *Task) createOrUpdate(task tekton.Task, clientset *client.ConfigSet) error {
	var taskObj *tekton.Task
	_, err := clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Create(&task)
	if k8sErrors.IsAlreadyExists(err) {
		taskObj, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Get(task.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		task.ObjectMeta.ResourceVersion = taskObj.GetResourceVersion()
		_, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Update(taskObj)
	}
	return err
}
