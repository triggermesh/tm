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

package revision

import (
	"context"
	"fmt"
	"time"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// GetTable converts k8s list instance into printable object
func (r *Revision) GetTable(list *servingv1.RevisionList) printer.Table {
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
		table.Rows = append(table.Rows, r.row(&item))
	}
	return table
}

func (r *Revision) row(item *servingv1.Revision) []string {
	name := item.Name
	namespace := item.Namespace
	// image := ""
	// if len(item.Spec.Containers) > 0 {
	// image = item.Spec.Containers[0].Image
	// }
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))
	ready := fmt.Sprintf("%v", item.IsReady())
	readyCondition := item.Status.GetCondition(servingv1.RevisionConditionReady)
	reason := ""
	if readyCondition != nil {
		reason = readyCondition.Reason
	}

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
func (r *Revision) List(clientset *client.ConfigSet) (*servingv1.RevisionList, error) {
	return clientset.Serving.ServingV1().Revisions(r.Namespace).List(context.Background(), metav1.ListOptions{})
}
