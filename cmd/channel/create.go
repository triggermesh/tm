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
	eventingApi "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Channel represents knative channel object
type Channel struct {
	Name        string
	Namespace   string
	Provisioner string
}

// Deploy knative eventing channel
func (c *Channel) Deploy(clientset *client.ConfigSet) error {
	channelObject := c.newObject(clientset)
	return c.createOrUpdate(channelObject, clientset)
}

func (c *Channel) newObject(clientset *client.ConfigSet) eventingApi.Channel {
	return eventingApi.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
		Spec: eventingApi.ChannelSpec{
			Provisioner: &corev1.ObjectReference{
				APIVersion: "eventing.knative.dev/v1alpha1",
				Kind:       "ClusterChannelProvisioner",
				Name:       c.Provisioner,
			},
		},
	}
}

func (c *Channel) createOrUpdate(channelObject eventingApi.Channel, clientset *client.ConfigSet) error {
	_, err := clientset.Eventing.EventingV1alpha1().Channels(c.Namespace).Create(&channelObject)
	if k8sErrors.IsAlreadyExists(err) {
		channel, err := clientset.Eventing.EventingV1alpha1().Channels(c.Namespace).Get(channelObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		channelObject.ObjectMeta.ResourceVersion = channel.GetResourceVersion()
		_, err = clientset.Eventing.EventingV1alpha1().Channels(c.Namespace).Update(&channelObject)
	}
	return err
}
