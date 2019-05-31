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

package taskrun

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knative/pkg/apis"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (tr *TaskRun) Deploy(clientset *client.ConfigSet) (string, error) {
	plr := pipelineresource.PipelineResource{
		Name:      tr.Resources,
		Namespace: tr.Namespace,
	}
	if _, err := plr.Get(clientset); err != nil {
		return "", err
	}
	image, err := tr.imageName(clientset)
	if err != nil {
		return "", err
	}
	taskRunObject := tr.newObject(image, clientset)
	result, err := clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Create(&taskRunObject)
	if err != nil {
		return "", err
	}
	tr.Name = result.GetName()
	if tr.Wait {
		fmt.Printf("waiting for %q ready state\n", result.Name)
		err = tr.wait(clientset)
	}
	return image, err
}

func (tr *TaskRun) newObject(registry string, clientset *client.ConfigSet) v1alpha1.TaskRun {
	return v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TaskRun",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: tr.Task + "-",
			Namespace:    tr.Namespace,
		},
		Spec: v1alpha1.TaskRunSpec{
			Trigger: v1alpha1.TaskTrigger{
				Type: v1alpha1.TaskTriggerTypeManual,
			},
			TaskRef: &v1alpha1.TaskRef{
				Kind:       "Task",
				APIVersion: "tekton.dev/v1alpha1",
				Name:       tr.Task,
			},
			Inputs: v1alpha1.TaskRunInputs{
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "sources",
						ResourceRef: v1alpha1.PipelineResourceRef{
							Name:       tr.Resources,
							APIVersion: "tekton.dev/v1alpha1",
						},
					},
				},
				Params: []v1alpha1.Param{
					{
						Name:  "registry",
						Value: registry,
					},
				},
			},
		},
	}
}

func (tr *TaskRun) imageName(clientset *client.ConfigSet) (string, error) {
	if len(tr.RegistrySecret) == 0 {
		return fmt.Sprintf("%s/%s/%s", tr.Registry, tr.Namespace, tr.Task), nil
	}
	secret, err := clientset.Core.CoreV1().Secrets(tr.Namespace).Get(tr.RegistrySecret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	data := secret.Data["config.json"]
	dec := json.NewDecoder(strings.NewReader(string(data)))
	var config registryAuths
	if err := dec.Decode(&config); err != nil {
		return "", err
	}
	if len(config.Auths) > 1 {
		return "", errors.New("credentials with multiple registries not supported")
	}
	for k, v := range config.Auths {
		if url, ok := gitlabEnv(); ok {
			return fmt.Sprintf("%s/%s", url, tr.Task), nil
		}
		return fmt.Sprintf("%s/%s/%s", k, v.Username, tr.Task), nil
	}
	return "", errors.New("empty registry credentials")
}

// hack to use correct username in image URL instead of "gitlab-ci-token" in Gitlab CI
func gitlabEnv() (string, bool) {
	return os.LookupEnv("CI_REGISTRY_IMAGE")
}

func (tr *TaskRun) wait(clientset *client.ConfigSet) error {
	res, err := clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", tr.Name),
	})
	if err != nil || res == nil {
		return fmt.Errorf("can't get watch interface: %s", err)
	}
	defer res.Stop()

	for {
		event := <-res.ResultChan()
		if event.Object == nil {
			res.Stop()
			if res, err = clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Watch(metav1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", tr.Name),
			}); err != nil || res == nil {
				return fmt.Errorf("can't restart watch interface: %s", err)
			}
			continue
		}
		taskrunEvent, ok := event.Object.(*v1alpha1.TaskRun)
		if !ok {
			continue
		}
		if taskrunEvent.IsDone() {
			return nil
		}
		for _, v := range taskrunEvent.Status.Conditions {
			if v.IsFalse() && v.Severity == apis.ConditionSeverityError {
				return errors.New(v.Message)
			}
		}
	}
}

func (tr *TaskRun) SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	taskrun, err := clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Get(tr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	taskrun.SetOwnerReferences([]metav1.OwnerReference{owner})
	_, err = clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Update(taskrun)
	return err
}
