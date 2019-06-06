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
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/task"

	"github.com/knative/pkg/apis"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	uploadDoneTrigger = ".uploadIsDone"
)

func (tr *TaskRun) Deploy(clientset *client.ConfigSet) (string, error) {
	var taskRunObject *v1alpha1.TaskRun
	if err := tr.SetupResources(clientset); err != nil {
		return "", fmt.Errorf("setup taskrun resources: %s", err)
	}
	if err := tr.checkPipelineResource(clientset); err != nil {
		return "", fmt.Errorf("checking taskrun pipelineresources: %s", err)
	}
	image, err := tr.imageName(clientset)
	if err != nil {
		return "", fmt.Errorf("composing image name: %s", err)
	}
	image = fmt.Sprintf("%s:%s", image, file.RandString(6))
	taskRunObject = tr.newTaskRun()
	taskRunObject.Spec.Inputs.Params = tr.getBuildArguments(image)

	taskRunObject, err = clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Create(taskRunObject)
	if err != nil {
		return "", fmt.Errorf("creating taskrun: %s", err)
	}
	tr.Name = taskRunObject.GetName()

	ownerRef := owner(taskRunObject)
	if tr.Task.Owned {
		tr.setTaskOwner(clientset, ownerRef)
	}
	if tr.PipelineResource.Owned {
		tr.setPipelineResourceOwner(clientset, ownerRef)
	}
	if file.IsLocal(tr.Source.URL) {
		pod, err := tr.taskPod(clientset)
		if err != nil {
			return "", fmt.Errorf("getting taskrun pod: %s", err)
		}
		sourceContainer, err := tr.sourceContainer(clientset, pod)
		if err != nil {
			return "", fmt.Errorf("waiting for source container: %s", err)
		}
		fmt.Printf("Uploading sources to %s\n", pod)
		if err := tr.injectSources(clientset, pod, sourceContainer); err != nil {
			return "", fmt.Errorf("injecting sources: %s", err)
		}
	}
	if tr.Wait {
		fmt.Printf("Waiting for taskrun %q ready state\n", taskRunObject.Name)
		err = tr.wait(clientset)
	}
	return image, err
}

func (tr *TaskRun) SetupResources(clientset *client.ConfigSet) error {
	newTask, err := tr.setupTask(clientset)
	if err != nil {
		return fmt.Errorf("task setup: %s", err)
	}
	if newTask != nil {
		tr.Task.Name = newTask.Name
		tr.Task.Owned = true
	}

	if tr.PipelineResource.Name == "" && file.IsGit(tr.Source.URL) {
		newPplRes, err := tr.setupPipelineresources(clientset)
		if err != nil {
			return fmt.Errorf("pipelineresource setup: %s", err)
		}
		tr.PipelineResource.Name = newPplRes.Name
		tr.PipelineResource.Owned = true
	}
	return nil
}

func (tr *TaskRun) setupPipelineresources(clientset *client.ConfigSet) (*v1alpha1.PipelineResource, error) {
	plr := pipelineresource.PipelineResource{
		Name:      tr.Name,
		Namespace: tr.Namespace,
		Source: pipelineresource.Git{
			URL:      tr.Source.URL,
			Revision: tr.Source.Revision,
		},
	}
	return plr.Deploy(clientset)
}

func (tr *TaskRun) setupTask(clientset *client.ConfigSet) (*v1alpha1.Task, error) {
	task := task.Task{
		Name:            tr.Task.Name,
		Namespace:       tr.Namespace,
		RegistrySecret:  tr.RegistrySecret,
		FromLocalSource: file.IsLocal(tr.Source.URL),
	}
	if taskObj, err := task.Get(clientset); err != nil {
		task.File = tr.Task.Name
		task.GenerateName = tr.Name + "-"
		return task.Deploy(clientset)
	} else if tr.RegistrySecret != "" {
		return task.Clone(clientset, taskObj)
	}
	return nil, nil
}

func (tr *TaskRun) setTaskOwner(clientset *client.ConfigSet, ownerRef metav1.OwnerReference) {
	task := task.Task{
		Name:      tr.Task.Name,
		Namespace: tr.Namespace,
	}
	if err := task.SetOwner(clientset, ownerRef); err != nil {
		if err = task.Delete(clientset); err != nil {
			fmt.Printf("Can't cleanup task: %s\n", err)
		}
	}
}

func (tr *TaskRun) setPipelineResourceOwner(clientset *client.ConfigSet, ownerRef metav1.OwnerReference) {
	plr := pipelineresource.PipelineResource{
		Name:      tr.PipelineResource.Name,
		Namespace: tr.Namespace,
	}
	if err := plr.SetOwner(clientset, ownerRef); err != nil {
		if err = plr.Delete(clientset); err != nil {
			fmt.Printf("Can't remove pipelineresource: %s\n", err)
		}
	}
}

func owner(taskRunObject *v1alpha1.TaskRun) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "tekton.dev/v1alpha1",
		Kind:       "TaskRun",
		Name:       taskRunObject.GetName(),
		UID:        taskRunObject.GetUID(),
	}
}

func (tr *TaskRun) checkPipelineResource(clientset *client.ConfigSet) error {
	if tr.PipelineResource.Name == "" {
		return nil
	}
	plr := pipelineresource.PipelineResource{
		Name:      tr.PipelineResource.Name,
		Namespace: tr.Namespace,
	}
	_, err := plr.Get(clientset)
	return err
}

func (tr *TaskRun) newTaskRun() *v1alpha1.TaskRun {
	name := tr.Name + "-"
	if tr.Name == "" {
		name = tr.Task.Name + "-"
	}
	taskrun := &v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TaskRun",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    tr.Namespace,
		},
		Spec: v1alpha1.TaskRunSpec{
			Trigger: v1alpha1.TaskTrigger{
				Type: v1alpha1.TaskTriggerTypeManual,
			},
			TaskRef: &v1alpha1.TaskRef{
				Kind:       "Task",
				APIVersion: "tekton.dev/v1alpha1",
				Name:       tr.Task.Name,
			},
			Inputs: v1alpha1.TaskRunInputs{},
		},
	}
	if tr.PipelineResource.Name != "" {
		taskrun.Spec.Inputs.Resources = []v1alpha1.TaskResourceBinding{
			{
				Name: "sources",
				ResourceRef: v1alpha1.PipelineResourceRef{
					Name:       tr.PipelineResource.Name,
					APIVersion: "tekton.dev/v1alpha1",
				},
			},
		}
	}
	return taskrun
}

func (tr *TaskRun) imageName(clientset *client.ConfigSet) (string, error) {
	if len(tr.RegistrySecret) == 0 {
		return fmt.Sprintf("%s/%s/%s", tr.Registry, tr.Namespace, tr.Name), nil
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
			return fmt.Sprintf("%s/%s", url, tr.Name), nil
		}
		return fmt.Sprintf("%s/%s/%s", k, v.Username, tr.Name), nil
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
			return tr.wait(clientset)
		}
		taskrun, ok := event.Object.(*v1alpha1.TaskRun)
		if !ok || taskrun == nil {
			continue
		}
		for _, v := range taskrun.Status.Conditions {
			if v.IsFalse() && v.Severity == apis.ConditionSeverityError {
				return errors.New(v.Message)
			}
		}
		if taskrun.IsDone() {
			return nil
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

func (tr *TaskRun) taskPod(clientset *client.ConfigSet) (string, error) {
	watch, err := clientset.Tekton.TektonV1alpha1().TaskRuns(tr.Namespace).Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + tr.Name,
	})
	if err != nil || watch == nil {
		return "", fmt.Errorf("Can't get taskrun %q watch interface", tr.Name)
	}
	defer watch.Stop()

	duration, err := time.ParseDuration(tr.Timeout)
	if err != nil {
		duration = 10 * time.Minute
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case event := <-watch.ResultChan():
			if event.Object == nil {
				return tr.taskPod(clientset)
			}
			res, ok := event.Object.(*v1alpha1.TaskRun)
			if !ok || res == nil {
				continue
			}
			if pod := res.Status.PodName; pod != "" {
				return pod, nil
			}
		case <-ticker.C:
			return "", fmt.Errorf("watch taskrun timeout")
		}
	}
	return "", nil
}

func (tr *TaskRun) sourceContainer(clientset *client.ConfigSet, podName string) (string, error) {
	watch, err := clientset.Core.CoreV1().Pods(tr.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + podName})
	if err != nil || watch == nil {
		return "", fmt.Errorf("can't get watch interface, please check taskrun status: %s", err)
	}
	defer watch.Stop()

	duration, err := time.ParseDuration(tr.Timeout)
	if err != nil {
		duration = 10 * time.Minute
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case event := <-watch.ResultChan():
			if event.Object == nil {
				return tr.sourceContainer(clientset, podName)
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok || pod == nil {
				continue
			}
			for _, v := range pod.Status.ContainerStatuses {
				if v.Name == "build-step-sources" {
					if v.State.Terminated != nil {
						// Looks like we got watch interface for "previous" service version
						return "", fmt.Errorf("taskrun container terminated")
					}
					if v.State.Running != nil {
						return v.Name, nil
					}
				}
			}
		case <-ticker.C:
			return "", fmt.Errorf("watch pod timeout")
		}
	}
}

func (tr *TaskRun) injectSources(clientset *client.ConfigSet, pod, container string) error {
	c := file.Copy{
		Pod:         pod,
		Namespace:   tr.Namespace,
		Container:   container,
		Source:      tr.Source.URL,
		Destination: path.Join("/home", path.Base(tr.Source.URL)),
	}
	if err := c.Upload(clientset); err != nil {
		return err
	}
	if _, _, err := c.RemoteExec(clientset, "touch "+uploadDoneTrigger, nil); err != nil {
		return err
	}
	return nil
}

func (tr *TaskRun) getBuildArguments(image string) []v1alpha1.Param {
	params := []v1alpha1.Param{
		{
			Name:  "IMAGE",
			Value: image,
		},
	}
	for k, v := range mapFromSlice(tr.Params) {
		params = append(params, v1alpha1.Param{
			Name: k, Value: v,
		})
	}
	return params
}

func mapFromSlice(slice []string) map[string]string {
	m := make(map[string]string)
	for _, s := range slice {
		t := regexp.MustCompile("[:=]").Split(s, 2)
		if len(t) != 2 {
			fmt.Printf("Can't parse argument slice %s\n", s)
			continue
		}
		m[t[0]] = t[1]
	}
	return m
}
