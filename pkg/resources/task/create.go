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

package task

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kind = "Task"
	api  = "tekton.dev/v1alpha1"
)

// Deploy accepts path (local or URL) to tekton Task manifest and installs it
func (t *Task) Deploy(clientset *client.ConfigSet) (*tekton.Task, error) {
	if !file.IsLocal(t.File) {
		path, err := file.Download(t.File)
		if err != nil {
			return nil, fmt.Errorf("task not found")
		}
		t.File = path
	}

	task, err := t.readYAML()
	if err != nil {
		return nil, err
	}

	task.SetNamespace(t.Namespace)
	if t.GenerateName != "" {
		task.SetName("")
		task.SetGenerateName(t.GenerateName)
	} else if t.Name != "" {
		task.SetName(t.Name)
	}

	if t.RegistrySecret != "" {
		t.setupEnv(task)
		t.setupVolume(task)
	}

	return t.createOrUpdate(task, clientset)
}

func (t *Task) Clone(clientset *client.ConfigSet, task *tekton.Task) (*tekton.Task, error) {
	task.Kind = kind
	task.APIVersion = api
	task.SetGenerateName(task.GetName() + "-")
	task.SetName("")
	task.SetResourceVersion("")
	if t.RegistrySecret != "" {
		t.setupEnv(task)
		t.setupVolume(task)
	}
	return t.createOrUpdate(task, clientset)
}

func (t *Task) readYAML() (*tekton.Task, error) {
	var res tekton.Task
	yamlFile, err := ioutil.ReadFile(t.File)
	if err != nil {
		return &res, err
	}
	err = yaml.Unmarshal(yamlFile, &res)
	return &res, err
}

func (t *Task) createOrUpdate(task *tekton.Task, clientset *client.ConfigSet) (*tekton.Task, error) {
	if task.TypeMeta.Kind != kind {
		return nil, fmt.Errorf("Object Kind mismatch: got %q, want %q", task.TypeMeta.Kind, kind)
	}
	if task.TypeMeta.APIVersion != api {
		return nil, fmt.Errorf("Object API mismatch: got %q, want %q", task.TypeMeta.APIVersion, api)
	}

	checkParams(task)

	if task.GetGenerateName() != "" {
		return clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Create(task)
	}

	taskObj, err := clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Create(task)
	if k8sErrors.IsAlreadyExists(err) {
		if taskObj, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Get(task.ObjectMeta.Name, metav1.GetOptions{}); err != nil {
			return nil, err
		}
		task.ObjectMeta.ResourceVersion = taskObj.GetResourceVersion()
		taskObj, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Update(taskObj)
	}
	return taskObj, err
}

func checkParams(task *tekton.Task) {
	var registry, sources bool
	for _, v := range task.Spec.Inputs.Params {
		if v.Name == "registry" {
			sources = true
			break
		}
	}
	for _, v := range task.Spec.Inputs.Resources {
		if v.Name == "sources" {
			registry = true
			break
		}
	}
	if !registry {
		fmt.Printf("Warning. Task %q does not have parameter \"registry\"", task.Name)
	}
	if !sources {
		fmt.Printf("Warning. Task %q does not have resource \"sources\"", task.Name)
	}
}

func (t *Task) SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	task, err := clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Get(t.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	task.SetOwnerReferences([]metav1.OwnerReference{owner})
	_, err = clientset.Tekton.TektonV1alpha1().Tasks(t.Namespace).Update(task)
	return err
}

func (t *Task) setupEnv(task *tekton.Task) {
	for i, container := range task.Spec.Steps {
		envs := append(container.Env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/" + t.RegistrySecret,
		})
		task.Spec.Steps[i].Env = envs
	}
}

func (t *Task) setupVolume(task *tekton.Task) {
	task.Spec.Volumes = []corev1.Volume{
		{
			Name: t.RegistrySecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: t.RegistrySecret,
				},
			},
		},
	}
	for i, step := range task.Spec.Steps {
		mounts := append(step.VolumeMounts, corev1.VolumeMount{
			Name:      t.RegistrySecret,
			MountPath: "/" + t.RegistrySecret,
			ReadOnly:  true,
		})
		task.Spec.Steps[i].VolumeMounts = mounts
	}
}
