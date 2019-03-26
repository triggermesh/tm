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

package pipeline

import (
	v1alpha1 "github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (pl *Pipeline) Deploy(clientset *client.ConfigSet) error {
	pipelineObject := pl.newObject(clientset)
	return pl.createOrUpdate(pipelineObject, clientset)
}

func (pl *Pipeline) newObject(clientset *client.ConfigSet) v1alpha1.Pipeline {
	return v1alpha1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pipeline",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pl.Name,
			Namespace: pl.Namespace,
		},
		Spec:   v1alpha1.PipelineSpec{},
		Status: v1alpha1.PipelineStatus{},
	}
}

func (pl *Pipeline) createOrUpdate(pipelineObject v1alpha1.Pipeline, clientset *client.ConfigSet) error {
	_, err := clientset.Tekton.TektonV1alpha1().Pipelines(pl.Namespace).Create(&pipelineObject)
	if k8sErrors.IsAlreadyExists(err) {
		pipeline, err := clientset.Tekton.TektonV1alpha1().Pipelines(pl.Namespace).Get(pipelineObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pipelineObject.ObjectMeta.ResourceVersion = pipeline.GetResourceVersion()
		_, err = clientset.Tekton.TektonV1alpha1().Pipelines(pl.Namespace).Update(&pipelineObject)
	}
	return err
}
