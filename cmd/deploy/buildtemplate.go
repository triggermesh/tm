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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/ghodss/yaml"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Buildtemplate struct {
	URL  string
	Path string
}

const (
	tmpPath = "/tmp"
)

func (b *Buildtemplate) DeployBuildTemplate(args []string, clientset *client.ClientSet) error {
	var bt buildv1alpha1.BuildTemplate
	var err error
	if len(b.URL) != 0 {
		fmt.Println("Downloading build template definition")
		if b.Path, err = downloadFile(b.URL); err != nil {
			return err
		}
	}
	if len(b.Path) == 0 {
		return errors.New("Empty path to buildtemplate yaml file")
	}
	if bt, err = readYaml(b.Path); err != nil {
		return err
	}
	// If argument is passed overwrite build template name
	if len(args) != 0 {
		bt.ObjectMeta.Name = args[0]
	}
	fmt.Printf("Creating \"%s\" build template\n", bt.ObjectMeta.Name)
	return createBuildTemplate(bt, clientset)
}

func readYaml(path string) (buildv1alpha1.BuildTemplate, error) {
	var res buildv1alpha1.BuildTemplate
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return res, err
	}
	err = yaml.Unmarshal(yamlFile, &res)
	return res, err
}

func createBuildTemplate(template buildv1alpha1.BuildTemplate, clientset *client.ClientSet) error {
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

func getBuildArguments(image string, buildArgs []string) ([]buildv1alpha1.ArgumentSpec, []buildv1alpha1.ParameterSpec) {
	args := []buildv1alpha1.ArgumentSpec{
		{
			Name:  "IMAGE",
			Value: image,
		},
	}

	for k, v := range getArgsFromSlice(buildArgs) {
		args = append(args, buildv1alpha1.ArgumentSpec{
			Name: k, Value: v,
		})
	}

	params := []buildv1alpha1.ParameterSpec{
		{
			Name: "IMAGE",
		},
	}

	for _, v := range args {
		params = append(params, buildv1alpha1.ParameterSpec{
			Name: v.Name,
		})
	}
	return args, params
}

func downloadFile(url string) (string, error) {
	path := tmpPath + "/" + time.Now().Format(time.RFC850)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}
