/*
Copyright (c) 2018 TriggerMesh, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deploy

import (
	"errors"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Build struct {
	Source        string
	Revision      string
	Step          string
	Command       []string
	Buildtemplate string
	Args          []string
	Image         string
}

func (b *Build) DeployBuild(args []string, clientset *client.ClientSet) error {
	build := buildv1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Build",
			APIVersion: "build.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: clientset.Namespace,
			CreationTimestamp: metav1.Time{
				time.Now(),
			},
		},
	}
	switch {
	case len(b.Buildtemplate) != 0:
		build.Spec = b.fromBuildtemplate(b.Buildtemplate, getArgsFromSlice(b.Args))
	case len(b.Step) != 0:
		steps := build.Spec.Steps
		existingBuild, err := clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).Get(args[0], metav1.GetOptions{})
		if err == nil {
			steps = existingBuild.Spec.Steps
		}
		build.Spec = b.fromBuildSteps(b.Step, b.Image, b.Command, b.Args, steps)
	default:
		return errors.New("Build steps or buildtemplate name must be specified")
	}

	build.Spec.Source = &buildv1alpha1.SourceSpec{
		Git: &buildv1alpha1.GitSourceSpec{
			Url:      b.Source,
			Revision: b.Revision,
		},
	}

	buildOld, err := clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).Get(build.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil {
		build.ObjectMeta.ResourceVersion = buildOld.ObjectMeta.ResourceVersion
		_, err = clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).Update(&build)
	} else if k8sErrors.IsNotFound(err) {
		_, err = clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).Create(&build)
	}
	return err
}

func (b *Build) fromBuildtemplate(name string, buildArgs map[string]string) buildv1alpha1.BuildSpec {
	// TODO add env variables support
	args := []buildv1alpha1.ArgumentSpec{}
	for k, v := range buildArgs {
		args = append(args, buildv1alpha1.ArgumentSpec{Name: k, Value: v})
	}

	return buildv1alpha1.BuildSpec{
		Template: &buildv1alpha1.TemplateInstantiationSpec{
			Name:      name,
			Arguments: args,
		},
	}
}

func (b *Build) fromBuildSteps(step, image string, command, args []string, steps []corev1.Container) buildv1alpha1.BuildSpec {
	steps = append(steps, corev1.Container{
		Name:    step,
		Image:   image,
		Args:    args,
		Command: command,
	})
	return buildv1alpha1.BuildSpec{Steps: steps}
}
