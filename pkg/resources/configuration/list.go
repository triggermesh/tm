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
	"fmt"
	"time"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// GetTable converts k8s list instance into printable object
func (cf *Configuration) GetTable(list *servingv1.ConfigurationList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			// "Image",
			"Age",
			"Ready",
			"Reason",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, cf.row(&item))
	}
	return table
}

func (cf *Configuration) row(item *servingv1.Configuration) []string {
	name := item.Name
	namespace := item.Namespace
	// image := ""
	// if len(item.Spec.Template.Spec.Containers) > 0 {
	// 	image = item.Spec.Template.Spec.Containers[0].Image
	// }
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))
	ready := fmt.Sprintf("%v", item.Status.IsReady())
	reason := item.Status.GetCondition(servingv1.ServiceConditionReady).Message

	row := []string{
		namespace,
		name,
		// image,
		age,
		ready,
		reason,
	}

	return row
}

// List returns k8s list object
func (cf *Configuration) List(clientset *client.ConfigSet) (*servingv1.ConfigurationList, error) {
	return clientset.Serving.ServingV1().Configurations(cf.Namespace).List(metav1.ListOptions{})
}
