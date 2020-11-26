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

package channel

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/triggermesh/tm/pkg/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingapi "knative.dev/eventing/pkg/apis/messaging/v1"
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

func (c *Channel) newObject(clientset *client.ConfigSet) messagingapi.InMemoryChannel {
	return messagingapi.InMemoryChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}
}

func (c *Channel) createOrUpdate(channelObject messagingapi.InMemoryChannel, clientset *client.ConfigSet) error {
	_, err := clientset.Eventing.MessagingV1().InMemoryChannels(c.Namespace).Create(context.Background(), &channelObject, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		channel, err := clientset.Eventing.MessagingV1().InMemoryChannels(c.Namespace).Get(context.Background(), channelObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		channelObject.ObjectMeta.ResourceVersion = channel.GetResourceVersion()
		_, err = clientset.Eventing.MessagingV1().InMemoryChannels(c.Namespace).Update(context.Background(), &channelObject, metav1.UpdateOptions{})
		return err
	}
	return err
}
