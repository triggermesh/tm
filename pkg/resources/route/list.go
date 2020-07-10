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

package route

import (
	"fmt"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// GetTable converts k8s list instance into printable object
func (rt *Route) GetTable(list *servingv1.RouteList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			"Url",
			"Ready",
			"Reason",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, rt.row(&item))
	}
	return table
}

func (rt *Route) row(item *servingv1.Route) []string {
	name := item.Name
	namespace := item.Namespace
	url := item.Status.URL.String()
	ready := fmt.Sprintf("%v", item.IsReady())
	readyCondition := item.Status.GetCondition(servingv1.RouteConditionReady)
	reason := ""
	if readyCondition != nil {
		reason = readyCondition.Reason
	}

	row := []string{
		namespace,
		name,
		url,
		ready,
		reason,
	}

	return row
}

// List returns k8s list object
func (rt *Route) List(clientset *client.ConfigSet) (*servingv1.RouteList, error) {
	return clientset.Serving.ServingV1().Routes(rt.Namespace).List(metav1.ListOptions{})
}
