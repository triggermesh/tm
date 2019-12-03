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

package configuration

import (
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func (c *Configuration) List(clientset *client.ConfigSet) (*servingv1alpha1.ConfigurationList, error) {
	return clientset.Serving.ServingV1alpha1().Configurations(c.Namespace).List(metav1.ListOptions{})
}
