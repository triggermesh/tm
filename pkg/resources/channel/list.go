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
	"time"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	messagingapi "knative.dev/eventing/pkg/apis/messaging/v1"
)

// GetTable converts k8s list instance into printable object
func (c *Channel) GetTable(list *messagingapi.InMemoryChannelList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			"Url",
			"Age",
			"Ready",
			"Reason",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, c.row(&item))
	}
	return table
}

func (c *Channel) row(item *messagingapi.InMemoryChannel) []string {
	name := item.Name
	namespace := item.Namespace
	url := item.Status.Address.URL.String()
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))
	ready := fmt.Sprintf("%v", item.Status.IsReady())
	readyCondition := item.Status.GetCondition(messagingapi.ChannelConditionReady)
	reason := ""
	if readyCondition != nil {
		reason = readyCondition.Reason
	}

	row := []string{
		namespace,
		name,
		url,
		age,
		ready,
		reason,
	}

	return row
}

// List returns list of knative build objects
func (c *Channel) List(clientset *client.ConfigSet) (*messagingapi.InMemoryChannelList, error) {
	return clientset.Eventing.MessagingV1().InMemoryChannels(c.Namespace).List(context.Background(), metav1.ListOptions{})
}
