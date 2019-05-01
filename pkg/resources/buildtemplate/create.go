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

package buildtemplate

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

// Deploy deploys knative buildtemplate either from local file or by its URL
func (b *Buildtemplate) Deploy(clientset *client.ConfigSet) (*buildv1alpha1.BuildTemplate, error) {
	var bt buildv1alpha1.BuildTemplate
	var err error

	if !file.IsLocal(b.File) {
		path, err := file.Download(b.File)
		if err != nil {
			return nil, fmt.Errorf("Buildtemplate %s: %s", b.File, err)
		}
		b.File = path
	}

	if bt, err = readYAML(b.File); err != nil {
		return nil, err
	}
	// If argument is passed overwrite build template name
	if len(b.Name) != 0 {
		bt.ObjectMeta.Name = b.Name
	}
	bt.ObjectMeta.Namespace = b.Namespace

	if len(b.RegistrySecret) != 0 {
		addSecretVolume(b.RegistrySecret, &bt)
		setEnvConfig(b.RegistrySecret, &bt)
	}

	return createBuildTemplate(bt, clientset)
}

func (bt *Buildtemplate) Clone(source buildv1alpha1.BuildTemplate, clientset *client.ConfigSet) (*buildv1alpha1.BuildTemplate, error) {
	source.SetGenerateName(bt.Name + "-")
	source.SetNamespace(bt.Namespace)
	source.SetOwnerReferences([]metav1.OwnerReference{})
	source.SetResourceVersion("")
	source.Kind = "BuildTemplate"
	if len(bt.RegistrySecret) != 0 {
		addSecretVolume(bt.RegistrySecret, &source)
		setEnvConfig(bt.RegistrySecret, &source)
	}
	return createBuildTemplate(source, clientset)
}

func addSecretVolume(registrySecret string, template *buildv1alpha1.BuildTemplate) {
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: registrySecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registrySecret,
				},
			},
		},
	}
	for i, step := range template.Spec.Steps {
		mounts := append(step.VolumeMounts, corev1.VolumeMount{
			Name:      registrySecret,
			MountPath: "/" + registrySecret,
			ReadOnly:  true,
		})
		template.Spec.Steps[i].VolumeMounts = mounts
	}
}

func setEnvConfig(registrySecret string, template *buildv1alpha1.BuildTemplate) {
	for i, step := range template.Spec.Steps {
		envs := append(step.Env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/" + registrySecret,
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

func createBuildTemplate(template buildv1alpha1.BuildTemplate, clientset *client.ConfigSet) (*buildv1alpha1.BuildTemplate, error) {
	if template.TypeMeta.Kind != "BuildTemplate" {
		return nil, errors.New("Can't create object, only BuildTemplate is allowed")
	}
	var hasImage bool
	for _, v := range template.Spec.Parameters {
		if v.Name == "IMAGE" {
			hasImage = true
			break
		}
	}
	if !hasImage {
		return nil, errors.New("Build template \"IMAGE\" parameter is missing")
	}
	btOld, err := clientset.Build.BuildV1alpha1().BuildTemplates(template.Namespace).Get(template.ObjectMeta.Name, metav1.GetOptions{})
	if err == nil {
		template.ObjectMeta.ResourceVersion = btOld.ObjectMeta.ResourceVersion
		return clientset.Build.BuildV1alpha1().BuildTemplates(template.Namespace).Update(&template)
	} else if k8sErrors.IsNotFound(err) {
		return clientset.Build.BuildV1alpha1().BuildTemplates(template.Namespace).Create(&template)
	}
	return nil, err
}
