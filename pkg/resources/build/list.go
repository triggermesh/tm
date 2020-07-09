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

package build

import (
	"fmt"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// GetTable converts k8s list instance into printable object
func (b *Build) GetTable(list *buildv1alpha1.BuildList) printer.Table {
	table := printer.Table{
		Headers: []string{
			"Namespace",
			"Name",
			"Source",
			"Buildtemplate",
			"Age",
			"Ready",
		},
		Rows: make([][]string, 0, len(list.Items)),
	}

	for _, item := range list.Items {
		table.Rows = append(table.Rows, b.row(&item))
	}
	return table
}

func (b *Build) row(item *buildv1alpha1.Build) []string {
	name := item.Name
	namespace := item.Namespace
	source := ""
	if item.Spec.Source != nil {
		if item.Spec.Source.Git != nil {
			source = item.Spec.Source.Git.Url
		}
	}
	buildtemplate := ""
	if item.Spec.Template != nil {
		buildtemplate = item.Spec.Template.Name
	}
	age := duration.HumanDuration(time.Since(item.GetCreationTimestamp().Time))
	ready := fmt.Sprintf("%t", item.Status.GetCondition(duckv1alpha1.ConditionSucceeded).IsTrue())

	row := []string{
		namespace,
		name,
		source,
		buildtemplate,
		age,
		ready,
	}

	return row
}

// List returns list of knative build objects
func (b *Build) List(clientset *client.ConfigSet) (*buildv1alpha1.BuildList, error) {
	return clientset.Build.BuildV1alpha1().Builds(b.Namespace).List(metav1.ListOptions{})
}
