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
	eventingv1alpha1 "github.com/knative/eventing-sources/pkg/apis/sources/v1alpha1"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	"github.com/triggermesh/tm/pkg/resources/service"
	"github.com/triggermesh/tm/pkg/resources/task"
	"gopkg.in/src-d/go-git.v4"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Push tries to read git configuration in current directory and if it succeeds
// tekton pipeline resource and task are being prepared to run "tm deploy" command.
// Corresponding TaskRun object which binds these pipelineresources and tasks
// is printed to stdout.
func Push(clientset *client.ConfigSet) error {
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

	pplr := pipelineresource.PipelineResource{
		Name:      project,
		Namespace: client.Namespace,
		Source: pipelineresource.Git{
			URL: url,
		},
	}
	if _, err := pplr.Deploy(clientset); err != nil {
		return err
	}

	t := task.Task{
		Name:      project,
		Namespace: client.Namespace,
	}

	if _, err := t.CreateOrUpdate(getTask(project, client.Namespace), clientset); err != nil {
		return err
	}

	tr := getTaskRun(project, client.Namespace)
	res, err := yaml.Marshal(tr)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project,
			Namespace: client.Namespace,
		},
		Data: map[string]string{
			"taskrun": string(res),
		},
	}

	if err := createOrUpdateCM(clientset, cm); err != nil {
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

	cs := getContainerSource(project, owner)
	if err := createOrUpdateCS(clientset, cs); err != nil {
		return err
	}

	tr.SetGenerateName("")
	tr.SetName(fmt.Sprintf("%s-%s", project, file.RandStringDNS(6)))
	res, err = yaml.Marshal(tr)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}

func getContainerSource(project, owner string) *eventingv1alpha1.ContainerSource {
	return &eventingv1alpha1.ContainerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerSource",
			APIVersion: "sources.eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project,
			Namespace: client.Namespace,
		},
		Spec: eventingv1alpha1.ContainerSourceSpec{
			Sink: &corev1.ObjectReference{
				Kind:       "Service",
				APIVersion: "serving.knative.dev/v1beta1",
				Name:       project + "-transceiver",
			},
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
					Value: "",
				}, {
					Name:  "EVENT_TYPE",
					Value: "commit",
				},
			},
		},
	}
}

func getTaskRun(taskName, namespace string) *tekton.TaskRun {
	return &tekton.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: taskName + "-",
			Namespace:    namespace,
		},
		Spec: tekton.TaskRunSpec{
			Inputs: tekton.TaskRunInputs{
				Resources: []tekton.TaskResourceBinding{
					{
						Name: "sources",
						ResourceRef: tekton.PipelineResourceRef{
							Name:       taskName,
							APIVersion: "tekton.dev/v1alpha1",
						},
					},
				},
			},
			TaskRef: &tekton.TaskRef{
				Name:       taskName,
				Kind:       "Task",
				APIVersion: "tekton.dev/v1alpha1",
			},
		},
	}
}

func getTask(name, namespace string) *tekton.Task {
	return &tekton.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "Task",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: tekton.TaskSpec{
			Inputs: &tekton.Inputs{
				Resources: []tekton.TaskResource{
					{
						ResourceDeclaration: tekton.ResourceDeclaration{
							Name: "sources",
							Type: tekton.PipelineResourceTypeGit,
						},
						OutputImageDir: "",
					},
				},
			},
			Steps: []tekton.Step{
				{
					corev1.Container{
						Name:    "deploy",
						Image:   "gcr.io/triggermesh/tm",
						Command: []string{"tm"},
						Args:    []string{"deploy", "-f", "/workspace/sources/"},
					},
				},
			},
		},
	}
}

func createOrUpdateCM(clientset *client.ConfigSet, cm *corev1.ConfigMap) error {
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

func createOrUpdateCS(clientset *client.ConfigSet, cs *eventingv1alpha1.ContainerSource) error {
	if _, err := clientset.EventSources.SourcesV1alpha1().ContainerSources(cs.Namespace).Create(cs); k8sErrors.IsAlreadyExists(err) {
		csObj, err := clientset.EventSources.SourcesV1alpha1().ContainerSources(cs.Namespace).Get(cs.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		cs.ObjectMeta.ResourceVersion = csObj.GetResourceVersion()
		if _, err := clientset.EventSources.SourcesV1alpha1().ContainerSources(cs.Namespace).Update(cs); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
