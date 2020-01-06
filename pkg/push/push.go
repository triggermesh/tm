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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	pipelines "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"gopkg.in/src-d/go-git.v4"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	githubSource "knative.dev/eventing-contrib/github/pkg/apis/sources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
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

	pipelineresource, err := json.Marshal(getPipelineResource(owner, project, url))
	if err != nil {
		return fmt.Errorf("pipelineresource marshaling: %s", err)
	}

	tr := getTaskRun(owner, project)
	taskrun, err := json.Marshal(tr)
	if err != nil {
		return fmt.Errorf("taskrun marshaling: %s", err)
	}

	triggertemplate := getTriggerTemplate(owner, project)
	triggerbinding := getTriggerBinding(owner, project)
	eventlistener := getEventListener(owner, project)
	githubsource := getGithubSource(owner, project)

	triggertemplate.Spec.ResourceTemplates = []triggers.TriggerResourceTemplate{
		{pipelineresource},
		{taskrun},
	}

	if err := createOrUpdateTriggerTemplate(triggertemplate, clientset); err != nil {
		return fmt.Errorf("creating triggertemplate: %s", err)
	}
	if err := createOrUpdateTriggerBinding(triggerbinding, clientset); err != nil {
		return fmt.Errorf("creating triggerbinding: %s", err)
	}
	if err := createOrUpdateEventListener(eventlistener, clientset); err != nil {
		return fmt.Errorf("creating eventlistener: %s", err)
	}
	if err := createOrUpdateGithubSource(githubsource, clientset); err != nil {
		return fmt.Errorf("creating githubsource: %s", err)
	}

	tr.Status = pipelines.TaskRunStatus{}
	tr.SetGenerateName("")
	tr.SetName(fmt.Sprintf("%s-%s-%s", owner, project, file.RandStringDNS(6)))
	res, err := yaml.Marshal(tr)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}

func getGithubSource(owner, project string) githubSource.GitHubSource {
	return githubSource.GitHubSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitHubSource",
			APIVersion: "sources.eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-source", owner, project),
			Namespace: client.Namespace,
		},
		Spec: githubSource.GitHubSourceSpec{
			OwnerAndRepository: fmt.Sprintf("%s/%s", owner, project),
			EventTypes:         []string{"push"},
			AccessToken: githubSource.SecretValueFromSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "githubsecret",
					},
					Key: "accessToken",
				},
			},
			SecretToken: githubSource.SecretValueFromSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "githubsecret",
					},
					Key: "secretToken",
				},
			},
			Sink: &duckv1beta1.Destination{
				Ref: &corev1.ObjectReference{
					Kind:       "EventListener",
					Name:       fmt.Sprintf("%s-%s-listener", owner, project),
					APIVersion: "tekton.dev/v1alpha1",
				},
			},
		},
	}
}

func getEventListener(owner, project string) triggers.EventListener {
	return triggers.EventListener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventListener",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-listener", owner, project),
			Namespace: client.Namespace,
		},
		Spec: triggers.EventListenerSpec{
			Triggers: []triggers.EventListenerTrigger{
				{
					Binding: &triggers.EventListenerBinding{
						Name: fmt.Sprintf("%s-%s-binding", owner, project),
					},
					Template: triggers.EventListenerTemplate{
						Name: fmt.Sprintf("%s-%s-template", owner, project),
					},
				},
			},
		},
	}
}

func getTriggerBinding(owner, project string) triggers.TriggerBinding {
	return triggers.TriggerBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TriggerBinding",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-binding", owner, project),
			Namespace: client.Namespace,
		},
		Spec: triggers.TriggerBindingSpec{
			Params: []pipelines.Param{
				{
					Name: "gitrevision",
					Value: pipelines.ArrayOrString{
						StringVal: "$(body.head_commit.id)",
						Type:      "string",
					},
				},
			},
		},
	}
}

func getTriggerTemplate(owner, project string) triggers.TriggerTemplate {
	return triggers.TriggerTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TriggerTemplate",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-template", owner, project),
			Namespace: client.Namespace,
		},
		Spec: triggers.TriggerTemplateSpec{
			Params: []pipelines.ParamSpec{
				{
					Name:        "gitrevision",
					Description: "The git revision",
					Default: &pipelines.ArrayOrString{
						StringVal: "master",
						Type:      "string",
					},
				},
			},
			ResourceTemplates: []triggers.TriggerResourceTemplate{},
		},
	}
}

func getPipelineResource(owner, project, url string) pipelines.PipelineResource {
	return pipelines.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineResource",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-resource", owner, project),
			Namespace: client.Namespace,
		},
		Spec: pipelines.PipelineResourceSpec{
			Type: pipelines.PipelineResourceTypeGit,
			Params: []pipelines.ResourceParam{
				{Name: "url", Value: url},
				{Name: "revision", Value: "$(params.gitrevision)"},
			},
		},
	}
}

func getTaskRun(owner, project string) *pipelines.TaskRun {
	return &pipelines.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", owner, project),
			Namespace:    client.Namespace,
		},
		Spec: pipelines.TaskRunSpec{
			Inputs: pipelines.TaskRunInputs{
				Resources: []pipelines.TaskResourceBinding{
					{
						PipelineResourceBinding: pipelines.PipelineResourceBinding{
							Name: "sources",
							ResourceRef: &pipelines.PipelineResourceRef{
								Name:       fmt.Sprintf("%s-%s-resource", owner, project),
								APIVersion: "tekton.dev/v1alpha1",
							},
						},
					},
				},
			},
			TaskRef: &pipelines.TaskRef{
				Name:       fmt.Sprintf("%s-%s-task", owner, project),
				Kind:       "Task",
				APIVersion: "tekton.dev/v1alpha1",
			},
		},
	}
}

func getTask(owner, project string) *pipelines.Task {
	return &pipelines.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "Task",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-task", owner, project),
			Namespace: client.Namespace,
		},
		Spec: pipelines.TaskSpec{
			Inputs: &pipelines.Inputs{
				Resources: []pipelines.TaskResource{
					{
						ResourceDeclaration: pipelines.ResourceDeclaration{
							Name: "sources",
							Type: pipelines.PipelineResourceTypeGit,
						},
					},
				},
			},
			Steps: []pipelines.Step{
				pipelines.Step{
					Container: corev1.Container{
						Name:    "deploy",
						Image:   "gcr.io/triggermesh/tm",
						Ports:   []corev1.ContainerPort{},
						Command: []string{"tm"},
						Args:    []string{"deploy", "-f", "/workspace/sources/"},
					},
				},
			},
		},
	}
}

func createOrUpdateTriggerTemplate(object triggers.TriggerTemplate, clientset *client.ConfigSet) error {
	_, err := clientset.TektonTriggers.TektonV1alpha1().TriggerTemplates(client.Namespace).Create(&object)
	if k8sErrors.IsAlreadyExists(err) {
		tt, err := clientset.TektonTriggers.TektonV1alpha1().TriggerTemplates(client.Namespace).Get(object.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		object.ObjectMeta.ResourceVersion = tt.GetResourceVersion()
		_, err = clientset.TektonTriggers.TektonV1alpha1().TriggerTemplates(client.Namespace).Update(&object)
		return err
	}
	return err
}

func createOrUpdateTriggerBinding(object triggers.TriggerBinding, clientset *client.ConfigSet) error {
	_, err := clientset.TektonTriggers.TektonV1alpha1().TriggerBindings(client.Namespace).Create(&object)
	if k8sErrors.IsAlreadyExists(err) {
		tb, err := clientset.TektonTriggers.TektonV1alpha1().TriggerBindings(client.Namespace).Get(object.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		object.ObjectMeta.ResourceVersion = tb.GetResourceVersion()
		_, err = clientset.TektonTriggers.TektonV1alpha1().TriggerBindings(client.Namespace).Update(&object)
		return err
	}
	return err
}

func createOrUpdateEventListener(object triggers.EventListener, clientset *client.ConfigSet) error {
	_, err := clientset.TektonTriggers.TektonV1alpha1().EventListeners(client.Namespace).Create(&object)
	if k8sErrors.IsAlreadyExists(err) {
		el, err := clientset.TektonTriggers.TektonV1alpha1().EventListeners(client.Namespace).Get(object.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		object.ObjectMeta.ResourceVersion = el.GetResourceVersion()
		_, err = clientset.TektonTriggers.TektonV1alpha1().EventListeners(client.Namespace).Update(&object)
		return err
	}
	return err
}

func createOrUpdateGithubSource(object githubSource.GitHubSource, clientset *client.ConfigSet) error {
	_, err := clientset.GithubSource.SourcesV1alpha1().GitHubSources(client.Namespace).Create(&object)
	if k8sErrors.IsAlreadyExists(err) {
		gh, err := clientset.GithubSource.SourcesV1alpha1().GitHubSources(client.Namespace).Get(object.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		object.ObjectMeta.ResourceVersion = gh.GetResourceVersion()
		_, err = clientset.GithubSource.SourcesV1alpha1().GitHubSources(client.Namespace).Update(&object)
		return err
	}
	return err
}
