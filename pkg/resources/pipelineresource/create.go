// Copyright 2018 TriggerMesh, Inc
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
	v1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (plr *PipelineResource) Deploy(clientset *client.ConfigSet) error {
	pipelineResourceObject := plr.newObject(clientset)
	return plr.createOrUpdate(pipelineResourceObject, clientset)
}

func (plr *PipelineResource) newObject(clientset *client.ConfigSet) v1alpha1.PipelineResource {
	return v1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pipeline",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      plr.Name,
			Namespace: plr.Namespace,
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type:         v1alpha1.PipelineResourceTypeGit,
			Params:       []v1alpha1.Param{},
			SecretParams: []v1alpha1.SecretParam{},
		},
		Status: v1alpha1.PipelineResourceStatus{},
	}
}

func (plr *PipelineResource) createOrUpdate(pipelineResourceObject v1alpha1.PipelineResource, clientset *client.ConfigSet) error {
	_, err := clientset.Tekton.TektonV1alpha1().PipelineResources(plr.Namespace).Create(&pipelineResourceObject)
	if k8sErrors.IsAlreadyExists(err) {
		pipeline, err := clientset.Tekton.TektonV1alpha1().PipelineResources(plr.Namespace).Get(pipelineResourceObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pipelineResourceObject.ObjectMeta.ResourceVersion = pipeline.GetResourceVersion()
		_, err = clientset.Tekton.TektonV1alpha1().PipelineResources(plr.Namespace).Update(&pipelineResourceObject)
	}
	return err
}
