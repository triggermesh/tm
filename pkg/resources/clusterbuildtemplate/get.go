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

package clusterbuildtemplate

import (
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get returns knative ClusterBuildTemplate object by its name
func (cbt *ClusterBuildtemplate) Get(clientset *client.ConfigSet) (*buildv1alpha1.ClusterBuildTemplate, error) {
	return clientset.Build.BuildV1alpha1().ClusterBuildTemplates().Get(cbt.Name, metav1.GetOptions{})
}

// Exist returns true if ClusterBuildTemplate with provided name is available
func Exist(clientset *client.ConfigSet, name string) bool {
	c := ClusterBuildtemplate{
		Name: name,
	}
	if _, err := c.Get(clientset); err == nil {
		return true
	}
	return false
}
