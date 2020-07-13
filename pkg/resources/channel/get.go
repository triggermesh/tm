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

package channel

import (
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingApi "knative.dev/eventing/pkg/apis/messaging/v1beta1"
)

// GetObject converts k8s object into printable structure
func (c *Channel) GetObject(service *messagingApi.InMemoryChannel) printer.Object {
	return printer.Object{
		Fields: map[string]interface{}{
			"Kind":              metav1.TypeMeta{}.Kind,
			"APIVersion":        metav1.TypeMeta{}.APIVersion,
			"Namespace":         metav1.ObjectMeta{}.Namespace,
			"Name":              metav1.ObjectMeta{}.Name,
			"CreationTimestamp": metav1.Time{},
			"Status":            messagingApi.InMemoryChannelStatus{},
		},
		K8sObject: service,
	}
}

// Get returns k8s object
func (c *Channel) Get(clientset *client.ConfigSet) (*messagingApi.InMemoryChannel, error) {
	return clientset.Eventing.MessagingV1beta1().InMemoryChannels(c.Namespace).Get(c.Name, metav1.GetOptions{})
}
