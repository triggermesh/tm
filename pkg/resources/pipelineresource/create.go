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
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy creates tekton PipelineResource with provided git URL
func (plr *PipelineResource) Deploy(clientset *client.ConfigSet) (*v1alpha1.PipelineResource, error) {
	pipelineResourceObject := plr.newObject(clientset)
	if client.Dry {
		return &pipelineResourceObject, nil
	}
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
			Type: v1alpha1.PipelineResourceTypeGit,
			Params: []v1alpha1.ResourceParam{
				{Name: "url", Value: plr.Source.URL},
				{Name: "revision", Value: plr.Source.Revision},
			},
		},
	}
}

func (plr *PipelineResource) createOrUpdate(pipelineResourceObject v1alpha1.PipelineResource, clientset *client.ConfigSet) (*v1alpha1.PipelineResource, error) {
	var pipeline *v1alpha1.PipelineResource
	res, err := clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).Create(&pipelineResourceObject)
	if k8serrors.IsAlreadyExists(err) {
		pipeline, err = clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).Get(pipelineResourceObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return res, err
		}
		pipelineResourceObject.ObjectMeta.ResourceVersion = pipeline.GetResourceVersion()
		res, err = clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).Update(&pipelineResourceObject)
	}
	return res, err
}

// SetOwner updates PipelineResource object with provided owner reference
func (plr *PipelineResource) SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	pplresource, err := clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).Get(plr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	pplresource.SetOwnerReferences([]metav1.OwnerReference{owner})
	_, err = clientset.TektonPipelines.TektonV1alpha1().PipelineResources(plr.Namespace).Update(pplresource)
	return err
}
