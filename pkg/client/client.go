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

package client

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	buildApi "github.com/knative/build/pkg/client/clientset/versioned"
	servingApi "github.com/knative/serving/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	confPath = "/.tm/config.json"
)

type ClientSet struct {
	Core      *kubernetes.Clientset
	Build     *buildApi.Clientset
	Serving   *servingApi.Clientset
	Namespace string
	Registry  string
}

type confStruct struct {
	Contexts []struct {
		Context struct {
			Cluster   string `json:"cluster"`
			Namespace string `json:"namespace"`
		} `json:"context"`
		Name string `json:"name"`
	} `json:"contexts"`
}

func username(cfgFile string) (string, error) {
	jsonFile, err := os.Open(cfgFile)
	if err != nil {
		return "", err
	}
	defer jsonFile.Close()

	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}
	if body, err = yaml.YAMLToJSON(body); err != nil {
		return "", err
	}

	var conf confStruct
	if err := yaml.Unmarshal(body, &conf); err != nil {
		return "", err
	}
	for _, v := range conf.Contexts {
		// TODO remove hardcoded cluster name
		if v.Context.Cluster == "triggermesh" {
			return v.Context.Namespace, nil
		}
	}
	return "default", nil
}

func NewClient(cfgFile, namespace, registry string) (ClientSet, error) {
	c := ClientSet{
		Namespace: namespace,
		Registry:  registry,
	}
	homeDir := "."
	if dir := os.Getenv("HOME"); dir != "" {
		homeDir = dir
	}
	tmHome := filepath.Dir(homeDir + confPath)
	if _, err := os.Stat(tmHome); os.IsNotExist(err) {
		if err := os.MkdirAll(tmHome, 0755); err != nil {
			log.Fatalln(err)
		}
	}

	if len(cfgFile) == 0 {
		cfgFile = homeDir + confPath
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", cfgFile); err != nil {
			if config, err = clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config"); err != nil {
				log.Fatalln("Can't read config file")
			}
			cfgFile = homeDir + "/.kube/config"
		}
		if len(namespace) == 0 {
			c.Namespace, err = username(cfgFile)
			if err != nil {
				return c, err
			}
		}
	}

	if c.Build, err = buildApi.NewForConfig(config); err != nil {
		return c, err
	}
	if c.Serving, err = servingApi.NewForConfig(config); err != nil {
		return c, err
	}
	if c.Core, err = kubernetes.NewForConfig(config); err != nil {
		return c, err
	}
	return c, nil
}
