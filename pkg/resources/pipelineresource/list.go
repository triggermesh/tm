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

package pipelineresource

import (
	"context"

	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetTable converts k8s list instance into printable object
func (plr *PipelineResource) GetTable(list *v1alpha1.PipelineResourceList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, plr.row(&item))
	}
	return table
}

func (plr *PipelineResource) row(item *v1alpha1.PipelineResource) []string {
	name := item.Name
	namespace := item.Namespace

	row := []string{
		namespace,
		name,
	}

	return row
}

// List returns k8s list object
func (plr *PipelineResource) List(clientset *client.ConfigSet) (*v1alpha1.PipelineResourceList, error) {
	return clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).List(context.Background(), metav1.ListOptions{})
}
