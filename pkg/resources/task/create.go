// Copyright 2018 TriggerMesh, Inc
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
	v1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *Task) Deploy(clientset *client.ConfigSet) error {
	taskObject := t.newObject(clientset)
	return t.createOrUpdate(taskObject, clientset)
}

func (t *Task) newObject(clientset *client.ConfigSet) v1alpha1.Task {
	return v1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
		},
		Spec: v1alpha1.TaskSpec{
			Steps: []corev1.Container{
				corev1.Container{
					Name:  t.Name,
					Image: "Foo",
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{Name: t.Name},
			},
		},
	}
}

func (t *Task) createOrUpdate(taskObject v1alpha1.Task, clientset *client.ConfigSet) error {
	_, err := clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Create(&taskObject)
	if k8sErrors.IsAlreadyExists(err) {
		task, err := clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Get(taskObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		taskObject.ObjectMeta.ResourceVersion = task.GetResourceVersion()
		_, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Update(&taskObject)
	}
	return err
}
