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

package taskrun

import (
	"fmt"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"knative.dev/pkg/apis"
)

// GetTable converts k8s list instance into printable object
func (tr *TaskRun) GetTable(list *v1beta1.TaskRunList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			"Age",
			"Succeeded",
			"Reason",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, tr.row(&item))
	}
	return table
}

func (tr *TaskRun) row(item *v1beta1.TaskRun) []string {
	name := item.Name
	namespace := item.Namespace
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))
	ready := fmt.Sprintf("%v", item.Status.GetCondition(apis.ConditionSucceeded).IsTrue())
	readyCondition := item.Status.GetCondition(apis.ConditionSucceeded)
	reason := ""
	if readyCondition != nil {
		reason = readyCondition.Reason
	}

	row := []string{
		namespace,
		name,
		age,
		ready,
		reason,
	}

	return row
}

// List returns k8s list object
func (tr *TaskRun) List(clientset *client.ConfigSet) (*v1beta1.TaskRunList, error) {
	return clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).List(metav1.ListOptions{})
}
