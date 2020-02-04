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
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func (tr *TaskRun) GetObject(taskrun *v1alpha1.TaskRun) printer.Object {
	return printer.Object{
		Fields: map[string]interface{}{
			"Kind":              metav1.TypeMeta{}.Kind,
			"APIVersion":        metav1.TypeMeta{}.APIVersion,
			"Namespace":         metav1.ObjectMeta{}.Namespace,
			"Name":              metav1.ObjectMeta{}.Name,
			"CreationTimestamp": metav1.Time{},
			"Spec":              v1alpha1.TaskRunSpec{},
			"Conditions":        duckv1beta1.Conditions{},
		},
		K8sObject: taskrun,
	}
}

func (tr *TaskRun) Get(clientset *client.ConfigSet) (*v1alpha1.TaskRun, error) {
	return clientset.TektonPipelines.TektonV1alpha1().TaskRuns(tr.Namespace).Get(tr.Name, metav1.GetOptions{})
}
