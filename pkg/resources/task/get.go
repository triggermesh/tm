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
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get return tekton Task object
func (t *Task) Get(clientset *client.ConfigSet) (*v1alpha1.Task, error) {
	return clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Get(t.Name, metav1.GetOptions{})
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
