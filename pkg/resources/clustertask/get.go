// Copyright 2020 TriggerMesh Inc.
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

package clustertask

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get returns tekton ClusterTask object by its name
func (ct *ClusterTask) Get(clientset *client.ConfigSet) (*v1beta1.ClusterTask, error) {
	return clientset.TektonTasks.TektonV1beta1().ClusterTasks().Get(ct.Name, metav1.GetOptions{})
}

// Exist returns true if ClusterTask with provided name is available
func Exist(clientset *client.ConfigSet, name string) bool {
	c := ClusterTask{
		Name: name,
	}
	if _, err := c.Get(clientset); err == nil {
		return true
	}
	return false
}
