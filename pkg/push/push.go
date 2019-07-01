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
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	"github.com/triggermesh/tm/pkg/resources/task"
	"gopkg.in/src-d/go-git.v4"
	corev1 "k8s.io/api/core/v1"
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

	res, err := yaml.Marshal(getTaskRun(project, client.Namespace))
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", res)
	return nil
}

func getTaskRun(taskName, namespace string) *tekton.TaskRun {
	return &tekton.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", taskName, file.RandStringDNS(6)),
			Namespace: namespace,
		},
		Spec: tekton.TaskRunSpec{
			Inputs: tekton.TaskRunInputs{
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "sources",
						ResourceRef: v1alpha1.PipelineResourceRef{
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
						Name: "sources",
						Type: tekton.PipelineResourceType("git"),
					},
				},
			},
			Steps: []corev1.Container{
				{
					Name:    "deploy",
					Image:   "gcr.io/triggermesh/tm",
					Command: []string{"tm"},
					Args:    []string{"deploy", "-f", "/workspace/sources/"},
				},
			},
		},
	}
}
