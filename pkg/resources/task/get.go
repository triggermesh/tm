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
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetObject converts k8s object into printable structure
func (t *Task) GetObject(task *v1beta1.Task) printer.Object {
	return printer.Object{
		Fields: map[string]interface{}{
			"Kind":              metav1.TypeMeta{}.Kind,
			"APIVersion":        metav1.TypeMeta{}.APIVersion,
			"Namespace":         metav1.ObjectMeta{}.Namespace,
			"Name":              metav1.ObjectMeta{}.Name,
			"CreationTimestamp": metav1.Time{},
			"Spec":              v1alpha1.TaskSpec{},
		},
		K8sObject: task,
	}
}

// Get return tekton Task object
func (t *Task) Get(clientset *client.ConfigSet) (*v1beta1.Task, error) {
	return clientset.TektonTasks.TektonV1beta1().Tasks(t.Namespace).Get(t.Name, metav1.GetOptions{})
}

// Exist returns true if Task with provided name is available in current namespace
func Exist(clientset *client.ConfigSet, name string) bool {
	t := Task{
		Name:      name,
		Namespace: client.Namespace,
	}
	if _, err := t.Get(clientset); err == nil {
		return true
	}
	return false
}
