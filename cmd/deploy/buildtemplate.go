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
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Buildtemplate contains information about knative buildtemplate definition
type Buildtemplate struct {
	Name           string
	File           string
	RegistrySecret string
}

// Deploy deploys knative buildtemplate either from local file or by its URL
func (b *Buildtemplate) Deploy(clientset *client.ConfigSet) (string, error) {
	var bt buildv1alpha1.BuildTemplate
	var err error

	if !file.IsLocal(b.File) {
		path, err := file.Download(b.File)
		if err != nil {
			return "", fmt.Errorf("Buildtemplate %s: %s", b.File, err)
		}
		b.File = path
	}

	if bt, err = readYAML(b.File); err != nil {
		return "", err
	}
	// If argument is passed overwrite build template name
	if len(b.Name) != 0 {
		bt.ObjectMeta.Name = b.Name
	}
	// fmt.Printf("Creating \"%s\" build template\n", bt.ObjectMeta.Name)

	if len(b.RegistrySecret) != 0 {
		b.addSecretVolume(&bt)
		b.setEnvConfig(&bt)
	}

	return bt.ObjectMeta.Name, createBuildTemplate(bt, clientset)
}

func (b *Buildtemplate) addSecretVolume(template *buildv1alpha1.BuildTemplate) {
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: b.RegistrySecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: b.RegistrySecret,
				},
			},
		},
	}
	for i, step := range template.Spec.Steps {
		mounts := append(step.VolumeMounts, corev1.VolumeMount{
			Name:      b.RegistrySecret,
			MountPath: "/" + b.RegistrySecret,
			ReadOnly:  true,
		})
		template.Spec.Steps[i].VolumeMounts = mounts
	}
}

func (b *Buildtemplate) setEnvConfig(template *buildv1alpha1.BuildTemplate) {
	for i, step := range template.Spec.Steps {
		envs := append(step.Env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/" + b.RegistrySecret,
		})
		template.Spec.Steps[i].Env = envs
	}
}

func readYAML(path string) (buildv1alpha1.BuildTemplate, error) {
	var res buildv1alpha1.BuildTemplate
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return res, err
	}
	err = yaml.Unmarshal(yamlFile, &res)
	return res, err
}

func createBuildTemplate(template buildv1alpha1.BuildTemplate, clientset *client.ConfigSet) error {
	if template.TypeMeta.Kind != "BuildTemplate" {
		return errors.New("Can't create object, only BuildTemplate is allowed")
	}
	var hasImage bool
	for _, v := range template.Spec.Parameters {
		if v.Name == "IMAGE" {
			hasImage = true
			break
		}
	}
	if !hasImage {
		return errors.New("Build template \"IMAGE\" parameter is missing")
	}
	btOld, err := clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Get(template.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil {
		template.ObjectMeta.ResourceVersion = btOld.ObjectMeta.ResourceVersion
		_, err = clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Update(&template)
	} else if k8sErrors.IsNotFound(err) {
		_, err = clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Create(&template)
	}
	return err
}

func getBuildArguments(image string, buildArgs []string) []buildv1alpha1.ArgumentSpec {
	args := []buildv1alpha1.ArgumentSpec{
		{
			Name:  "IMAGE",
			Value: image,
		},
	}

	for k, v := range mapFromSlice(buildArgs) {
		args = append(args, buildv1alpha1.ArgumentSpec{
			Name: k, Value: v,
		})
	}

	return args
}
