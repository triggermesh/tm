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

	"github.com/ghodss/yaml"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/clustertask"
	"github.com/triggermesh/tm/pkg/resources/pipelineresource"
	"github.com/triggermesh/tm/pkg/resources/task"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	tektonAPI         = "tekton.dev/v1beta1"
	taskRunKind       = "TaskRun"
	taskKind          = "Task"
	clusterTaskKind   = "ClusterTask"
	uploadDoneTrigger = ".uploadIsDone"
)

// Deploy prepares and verifies tekton resources (Task and PipelineResource) required for TaskRun,
// creates TaskRun object and optionally waits for its result.
// Deploy function returns resulting image URL and build error.
func (tr *TaskRun) Deploy(clientset *client.ConfigSet) (string, error) {
	if tr.Name == "" {
		return "", fmt.Errorf("taskrun name cannot be empty")
	}
	if tr.Task.Name == "" {
		return "", fmt.Errorf("task name cannot be empty")
	}
	if !client.Dry {
		if err := tr.prepareTask(clientset); err != nil {
			return "", fmt.Errorf("setup task: %s", err)
		}
		if err := tr.preparePipelineresources(clientset); err != nil {
			return "", fmt.Errorf("setup pipelineresource: %s", err)
		}
	}
	if err := tr.checkPipelineResource(clientset); err != nil {
		return "", fmt.Errorf("pipelineresource %q not found", tr.PipelineResource.Name)
	}
	image, err := tr.imageName(clientset)
	if err != nil {
		return "", fmt.Errorf("composing image name: %s", err)
	}
	image = fmt.Sprintf("%s:%s", image, file.RandString(6))
	clientset.Log.Debugf("taskrun \"%s/%s\" output image will be %q\n", tr.Namespace, tr.Name, image)
	taskRunObject := tr.newTaskRun()
	taskRunObject.Spec.Params = tr.getBuildArguments(image)

	if file.IsLocal(tr.Function.Path) {
		if file.IsDir(tr.Function.Path) {
			tr.Function.Path = path.Clean(tr.Function.Path)
		} else {
			tr.Params = append(tr.Params, "HANDLER="+path.Base(tr.Function.Path))
			tr.Function.Path = path.Clean(path.Dir(tr.Function.Path))
		}
		clientset.Log.Debugf("function path is %q\n", tr.Function.Path)
	}

	if client.Dry {
		var taskObj []byte
		if client.Output == "yaml" {
			taskObj, err = yaml.Marshal(taskRunObject)
		} else {
			taskObj, err = json.MarshalIndent(taskRunObject, "", " ")
		}
		return string(taskObj), err
	}

	taskRunObject, err = clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).Create(taskRunObject)
	if err != nil {
		return "", fmt.Errorf("creating taskrun: %s", err)
	}
	tr.Name = taskRunObject.GetName()
	clientset.Log.Debugf("taskrun \"%s/%s\" created\n", tr.Namespace, tr.Name)

	task := task.Task{
		Name:      tr.Task.Name,
		Namespace: tr.Namespace,
	}

	ownerRef := owner(taskRunObject)
	if tr.Task.Owned {
		err = task.SetOwner(clientset, ownerRef)
		if err != nil {
			if err := task.Delete(clientset); err != nil {
				clientset.Log.Errorf("Can't cleanup task: %s\n", err)
			}
			return "", err
		}
	}
	if tr.PipelineResource.Owned {
		clientset.Log.Debugf("setting pipelineresource owner\n")
		tr.setPipelineResourceOwner(clientset, ownerRef)
	}
	if file.IsLocal(tr.Function.Path) {
		pod, err := tr.taskPod(clientset)
		if err != nil {
			return "", fmt.Errorf("getting taskrun pod: %s", err)
		}
		sourceContainer, err := tr.sourceContainer(clientset, pod)
		if err != nil {
			return "", fmt.Errorf("waiting for source container: %s", err)
		}
		clientset.Log.Infof("Uploading %q to %s\n", tr.Function.Path, pod)
		if err := tr.injectSources(clientset, pod, sourceContainer); err != nil {
			return "", fmt.Errorf("injecting sources: %s", err)
		}
	}
	if tr.Wait {
		clientset.Log.Infof("Waiting for taskrun %q ready state\n", taskRunObject.Name)
		if err = tr.wait(clientset); err != nil {
			return image, fmt.Errorf("taskrun %q deployment failed: %s", tr.Name, err)
		}
	}
	return image, err
}

func (tr *TaskRun) prepareTask(clientset *client.ConfigSet) error {
	newTask, err := tr.setupTask(clientset)
	if err != nil {
		return fmt.Errorf("task %q setup: %s", tr.Task.Name, err)
	}
	if newTask != nil {
		tr.Task.Name = newTask.Name
		tr.Task.Owned = true
	}
	return nil
}

func (tr *TaskRun) preparePipelineresources(clientset *client.ConfigSet) error {
	if tr.PipelineResource.Name == "" && file.IsGit(tr.Function.Path) {
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
			URL:      tr.Function.Path,
			Revision: tr.Function.Revision,
		},
	}
	return plr.Deploy(clientset)
}

func (tr *TaskRun) setupTask(clientset *client.ConfigSet) (*v1beta1.Task, error) {
	task := task.Task{
		Name:            tr.Task.Name,
		Namespace:       tr.Namespace,
		FromLocalSource: file.IsLocal(tr.Function.Path),
	}
	taskObj, err := task.Get(clientset)
	if err != nil {
		clusterTask := clustertask.ClusterTask{
			Name: task.Name,
		}
		clustertaskObj, err := clusterTask.Get(clientset)
		if err != nil {
			task.File = tr.Task.Name
			task.GenerateName = tr.Name + "-"
			return task.Deploy(clientset)
		}
		taskObj.Spec = clustertaskObj.Spec
		taskObj.TypeMeta = clustertaskObj.TypeMeta
		taskObj.ObjectMeta = clustertaskObj.ObjectMeta
		tr.Task.ClusterScope = true
	}
	if clientset.Registry.Secret != "" || task.FromLocalSource {
		tr.Task.ClusterScope = false
		clientset.Log.Debugf("cloning task to a new object \"%s/%s\"\n", task.Namespace, task.Name)
		return task.Clone(clientset, taskObj)
	}
	return nil, nil
}

func (tr *TaskRun) setPipelineResourceOwner(clientset *client.ConfigSet, ownerRef metav1.OwnerReference) {
	plr := pipelineresource.PipelineResource{
		Name:      tr.PipelineResource.Name,
		Namespace: tr.Namespace,
	}
	if err := plr.SetOwner(clientset, ownerRef); err != nil {
		if err = plr.Delete(clientset); err != nil {
			clientset.Log.Errorf("Can't remove pipelineresource: %s\n", err)
		}
	}
}

func owner(taskRunObject *v1beta1.TaskRun) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "tekton.dev/v1beta1",
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

func (tr *TaskRun) newTaskRun() *v1beta1.TaskRun {
	// root := int64(0)
	name := tr.Name + "-"
	if tr.Name == "" {
		name = tr.Task.Name + "-"
	}
	taskref := &v1beta1.TaskRef{
		Kind:       taskKind,
		APIVersion: tektonAPI,
		Name:       tr.Task.Name,
	}
	if tr.Task.ClusterScope {
		taskref.Kind = clusterTaskKind
	}
	taskrun := &v1beta1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       taskRunKind,
			APIVersion: tektonAPI,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    tr.Namespace,
		},
		Spec: v1beta1.TaskRunSpec{
			TaskRef:   taskref,
			Resources: &v1beta1.TaskRunResources{},
			// Inputs:  &v1alpha1.TaskRunInputs{},
			// PodTemplate: v1alpha1.PodTemplate{
			// SecurityContext: &corev1.PodSecurityContext{
			// RunAsUser: &root,
			// },
			// },
		},
	}
	if tr.PipelineResource.Name != "" {
		taskrun.Spec.Resources.Inputs = []v1beta1.TaskResourceBinding{
			{
				PipelineResourceBinding: v1beta1.PipelineResourceBinding{
					Name: "sources",
					ResourceRef: &v1beta1.PipelineResourceRef{
						Name:       tr.PipelineResource.Name,
						APIVersion: "tekton.dev/v1beta1",
					},
				},
			},
		}
	}
	return taskrun
}

func (tr *TaskRun) imageName(clientset *client.ConfigSet) (string, error) {
	if len(clientset.Registry.Secret) == 0 {
		return fmt.Sprintf("%s/%s/%s", clientset.Registry.Host, tr.Namespace, tr.Name), nil
	}
	secret, err := clientset.Core.CoreV1().Secrets(tr.Namespace).Get(clientset.Registry.Secret, metav1.GetOptions{})
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
	trWatchInterface, err := clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", tr.Name),
	})
	if err != nil || trWatchInterface == nil {
		return fmt.Errorf("can't get watch interface: %s", err)
	}
	defer trWatchInterface.Stop()

	for {
		event := <-trWatchInterface.ResultChan()
		if event.Object == nil {
			return tr.wait(clientset)
		}
		taskrun, ok := event.Object.(*v1beta1.TaskRun)
		if !ok || taskrun == nil {
			continue
		}
		if clientset.Log.IsDebug() {
			clientset.Log.Debugf("got new event:\n")
			for _, v := range taskrun.Status.Conditions {
				clientset.Log.Debugf(" condition: %q, status: %q, message: %q\n", v.Type, v.Status, v.Message)
			}
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

// SetOwner updates TaskRun object with provided owner reference
func (tr *TaskRun) SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	taskrun, err := clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).Get(tr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	clientset.Log.Debugf("setting taskrun \"%s/%s\" owner to %s/%s\n", taskrun.GetNamespace(), taskrun.GetName(), owner.Kind, owner.Name)
	taskrun.SetOwnerReferences([]metav1.OwnerReference{owner})
	_, err = clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).Update(taskrun)
	return err
}

func (tr *TaskRun) taskPod(clientset *client.ConfigSet) (string, error) {
	watch, err := clientset.TektonTasks.TektonV1beta1().TaskRuns(tr.Namespace).Watch(metav1.ListOptions{
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
			res, ok := event.Object.(*v1beta1.TaskRun)
			if !ok || res == nil {
				continue
			}
			status := res.Status.GetCondition(apis.ConditionType("Succeeded"))
			if status != nil && status.IsFalse() {
				return "", fmt.Errorf("taskrun failed: %s", status.Message)
			}
			if pod := res.Status.PodName; pod != "" {
				return pod, nil
			}
		case <-ticker.C:
			return "", fmt.Errorf("watch taskrun timeout")
		}
	}
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
				if strings.HasSuffix(v.Name, "sources-receiver") {
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
		Source:      tr.Function.Path,
		Destination: path.Join("/home", path.Base(tr.Function.Path)),
	}
	if err := c.Upload(clientset); err != nil {
		return err
	}
	clientset.Log.Debugf("creating upload completion flag\n")
	if _, _, err := c.RemoteExec(clientset, "touch "+uploadDoneTrigger, nil); err != nil {
		return err
	}
	return nil
}

func (tr *TaskRun) getBuildArguments(image string) []v1beta1.Param {
	params := []v1beta1.Param{
		{
			Name: "IMAGE",
			Value: v1beta1.ArrayOrString{
				Type:      v1beta1.ParamTypeString,
				StringVal: image,
			},
		},
	}
	for k, v := range mapFromSlice(tr.Params) {
		params = append(params, v1beta1.Param{
			Name: k,
			Value: v1beta1.ArrayOrString{
				Type:      v1beta1.ParamTypeString,
				StringVal: v,
			},
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
