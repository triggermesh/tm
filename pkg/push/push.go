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

package push

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	"github.com/triggermesh/tm/pkg/resources/service"
	"github.com/triggermesh/tm/pkg/resources/task"
	"gopkg.in/src-d/go-git.v4"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// Push tries to read git configuration in current directory and if it succeeds
// tekton pipeline resource and task are being prepared to run "tm deploy" command.
// Corresponding TaskRun object which binds these pipelineresources and tasks
// is printed to stdout.
func Push(clientset *client.ConfigSet, token string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	remote, err := repo.Remote("origin")
	if err != nil {
		return err
	}
	if remote == nil {
		return fmt.Errorf("nil remote")
	}
	if len(remote.Config().URLs) == 0 {
		return fmt.Errorf("no remote URLs")
	}

	url := remote.Config().URLs[0]
	if prefix := strings.Index(url, "@"); prefix != 0 {
		url = strings.ReplaceAll(url[prefix+1:], ":", "/")
		url = strings.TrimRight(url, ".git")
	}

	url = fmt.Sprintf("https://%s", url)
	parts := strings.Split(url, "/")
	project := parts[len(parts)-1]
	owner := parts[len(parts)-2]

	pipelineResourceObj := pipelineresource.PipelineResource{
		Name:      project,
		Namespace: client.Namespace,
		Source: pipelineresource.Git{
			URL: url,
		},
	}
	if _, err := pipelineResourceObj.Deploy(clientset); err != nil {
		return err
	}

	taskObj := task.Task{
		Name:      project,
		Namespace: client.Namespace,
	}

	if _, err := taskObj.CreateOrUpdate(getTask(project, client.Namespace), clientset); err != nil {
		return err
	}

	taskrunObj := getTaskRun(project, client.Namespace)
	taskrunYAML, err := yaml.Marshal(taskrunObj)
	if err != nil {
		return err
	}

	configmapObj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project,
			Namespace: client.Namespace,
		},
		Data: map[string]string{
			"taskrun": string(taskrunYAML),
		},
	}

	if err := createOrUpdateConfigmap(clientset, configmapObj); err != nil {
		return err
	}

	kservice := service.Service{
		Name:      project + "-transceiver",
		Namespace: client.Namespace,
		Source:    "docker.io/triggermesh/transceiver",
		Env: []string{
			"TASKRUN_CONFIGMAP=" + project,
			"NAMESPACE=" + client.Namespace,
		},
	}

	if _, err := kservice.Deploy(clientset); err != nil {
		return err
	}

	containersource := getContainerSource(project, owner, token)
	if err := createOrUpdateContainersource(clientset, containersource); err != nil {
		return err
	}

	taskrunObj.SetGenerateName("")
	taskrunObj.SetName(fmt.Sprintf("%s-%s", project, file.RandStringDNS(6)))
	taskrunYAML, err = yaml.Marshal(taskrunObj)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", taskrunYAML)
	return nil
}

func getContainerSource(project, owner, token string) *sourcesv1alpha2.ContainerSource {
	return &sourcesv1alpha2.ContainerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerSource",
			APIVersion: "sources.eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project,
			Namespace: client.Namespace,
		},
		Spec: sourcesv1alpha2.ContainerSourceSpec{
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						Kind:       "Service",
						APIVersion: "serving.knative.dev/v1beta1",
						Name:       project + "-transceiver",
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "user-container",
							Image: "triggermesh/github-third-party-source",
							Env: []corev1.EnvVar{
								{
									Name:  "OWNER",
									Value: owner,
								}, {
									Name:  "REPOSITORY",
									Value: project,
								}, {
									Name:  "TOKEN",
									Value: token,
								}, {
									Name:  "EVENT_TYPE",
									Value: "commit",
								},
							},
						},
					},
				},
			},
		},
	}
}

func getTaskRun(taskName, namespace string) *tekton.TaskRun {
	return &tekton.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1beta1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: taskName + "-",
			Namespace:    namespace,
		},
		Spec: tekton.TaskRunSpec{
			Resources: &tekton.TaskRunResources{
				Inputs: []tekton.TaskResourceBinding{
					{
						PipelineResourceBinding: tekton.PipelineResourceBinding{
							Name: "url",
							ResourceRef: &tekton.PipelineResourceRef{
								Name:       taskName,
								APIVersion: "tekton.dev/v1beta1",
							},
						},
					},
				},
			},
			TaskRef: &tekton.TaskRef{
				Name:       taskName,
				Kind:       "Task",
				APIVersion: "tekton.dev/v1beta1",
			},
		},
	}
}

func getTask(name, namespace string) *tekton.Task {
	return &tekton.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1beta1",
			Kind:       "Task",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: tekton.TaskSpec{
			Resources: &tekton.TaskResources{
				Inputs: []tekton.TaskResource{
					{
						ResourceDeclaration: tekton.ResourceDeclaration{
							Name: "url",
							Type: tekton.PipelineResourceTypeGit,
						},
					},
				},
			},
			Steps: []tekton.Step{
				{
					Container: corev1.Container{
						Name:    "deploy",
						Image:   "gcr.io/triggermesh/tm",
						Command: []string{"tm"},
						// TODO: move sources from this illogical path; update pipelineresource/create.go
						Args: []string{"deploy", "-f", "/workspace/url/", "--wait"},
					},
				},
			},
		},
	}
}

func createOrUpdateConfigmap(clientset *client.ConfigSet, cm *corev1.ConfigMap) error {
	if _, err := clientset.Core.CoreV1().ConfigMaps(cm.Namespace).Create(cm); k8sErrors.IsAlreadyExists(err) {
		cmObj, err := clientset.Core.CoreV1().ConfigMaps(cm.Namespace).Get(cm.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		cm.ObjectMeta.ResourceVersion = cmObj.GetResourceVersion()
		if _, err := clientset.Core.CoreV1().ConfigMaps(cm.Namespace).Update(cm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func createOrUpdateContainersource(clientset *client.ConfigSet, cs *sourcesv1alpha2.ContainerSource) error {
	if _, err := clientset.Eventing.SourcesV1alpha2().ContainerSources(cs.Namespace).Create(cs); k8sErrors.IsAlreadyExists(err) {
		csObj, err := clientset.Eventing.SourcesV1alpha2().ContainerSources(cs.Namespace).Get(cs.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		cs.ObjectMeta.ResourceVersion = csObj.GetResourceVersion()
		if _, err := clientset.Eventing.SourcesV1alpha2().ContainerSources(cs.Namespace).Update(cs); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
