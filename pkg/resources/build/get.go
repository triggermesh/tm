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
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/printer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetObject converts k8s object into printable structure
func (b *Build) GetObject(build *buildv1alpha1.Build) printer.Object {
	return printer.Object{
		Fields: map[string]interface{}{
			"Kind":              metav1.TypeMeta{}.Kind,
			"APIVersion":        metav1.TypeMeta{}.APIVersion,
			"Namespace":         metav1.ObjectMeta{}.Namespace,
			"Name":              metav1.ObjectMeta{}.Name,
			"CreationTimestamp": metav1.Time{},
			"Source":            &buildv1alpha1.SourceSpec{},
			"Template":          &buildv1alpha1.TemplateInstantiationSpec{},
			"Conditions":        duckv1alpha1.Conditions{},
		},
		K8sObject: build,
	}
}

// Get returns k8s object
func (b *Build) Get(clientset *client.ConfigSet) (*buildv1alpha1.Build, error) {
	return clientset.Build.BuildV1alpha1().Builds(b.Namespace).Get(b.Name, metav1.GetOptions{})
}
