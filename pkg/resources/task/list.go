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

package task

import (
	"context"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// GetTable converts k8s list instance into printable object
func (t *Task) GetTable(list *v1beta1.TaskList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			"Age",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, t.row(&item))
	}
	return table
}

func (t *Task) row(item *v1beta1.Task) []string {
	name := item.Name
	namespace := item.Namespace
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))

	row := []string{
		namespace,
		name,
		age,
	}

	return row
}

// List returns k8s list object
func (t *Task) List(clientset *client.ConfigSet) (*v1beta1.TaskList, error) {
	return clientset.TektonTasks.TektonV1beta1().Tasks(t.Namespace).List(context.Background(), metav1.ListOptions{})
}
