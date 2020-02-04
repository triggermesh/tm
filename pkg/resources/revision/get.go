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

package revision

import (
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func (r *Revision) GetObject(revision *servingv1.Revision) printer.Object {
	return printer.Object{
		Fields: map[string]interface{}{
			"Kind":       metav1.TypeMeta{}.Kind,
			"APIVersion": metav1.TypeMeta{}.APIVersion,
			"Namespace":  metav1.ObjectMeta{}.Namespace,
			"Name":       metav1.ObjectMeta{}.Name,
			"Containers": []corev1.Container{},
			"Conditions": duckv1.Conditions{},
		},
		K8sObject: revision,
	}
}

func (r *Revision) Get(clientset *client.ConfigSet) (*servingv1.Revision, error) {
	return clientset.Serving.ServingV1().Revisions(r.Namespace).Get(r.Name, metav1.GetOptions{})
}
