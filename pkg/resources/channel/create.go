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
	"fmt"

	"github.com/ghodss/yaml"
	messagingApi "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy knative eventing channel
func (c *Channel) Deploy(clientset *client.ConfigSet) error {
	channelObject := c.newObject(clientset)
	if client.Dry {
		res, err := yaml.Marshal(channelObject)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", res)
		return nil
	}
	return c.createOrUpdate(channelObject, clientset)
}

func (c *Channel) newObject(clientset *client.ConfigSet) messagingApi.InMemoryChannel {
	return messagingApi.InMemoryChannel{
		TypeMeta: metav1.TypeMeta{
			Kind:       c.Kind,
			APIVersion: "messaging.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}
}

func (c *Channel) createOrUpdate(channelObject messagingApi.InMemoryChannel, clientset *client.ConfigSet) error {
	_, err := clientset.Eventing.MessagingV1alpha1().InMemoryChannels(c.Namespace).Create(&channelObject)
	if k8sErrors.IsAlreadyExists(err) {
		channel, err := clientset.Eventing.MessagingV1alpha1().InMemoryChannels(c.Namespace).Get(channelObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		channelObject.ObjectMeta.ResourceVersion = channel.GetResourceVersion()
		_, err = clientset.Eventing.MessagingV1alpha1().InMemoryChannels(c.Namespace).Update(&channelObject)
		return err
	}
	return err
}
