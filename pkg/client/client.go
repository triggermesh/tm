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

	// gcp package is required for kubectl configs with GCP auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	confPath = "/.tm/config.json"
)

// ConfigSet contains different information that may be needed by underlying functions
type ConfigSet struct {
	Core    *kubernetes.Clientset
	Build   *buildApi.Clientset
	Serving *servingApi.Clientset

	Config *rest.Config

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

// NewClient returns ConfigSet created from available configuration file or from in-cluster environment
func NewClient(cfgFile, namespace, registry string) (ConfigSet, error) {
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

	kubeconfig := os.Getenv("KUBECONFIG")
	if len(cfgFile) != 0 {
		// using config file passed with --config argument
	} else if _, err := os.Stat(homeDir + "/.tm/config.json"); err == nil {
		cfgFile = homeDir + "/.tm/config.json"
	} else if _, err := os.Stat(kubeconfig); err == nil {
		cfgFile = kubeconfig
	} else {
		cfgFile = homeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err == nil && len(namespace) == 0 {
		if namespace, err = username(cfgFile); err != nil {
			return ConfigSet{}, err
		}
	} else if err != nil {
		log.Printf("%s, falling back to in-cluster configuration\n", err)
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalln("Can't read config file")
	}

	c := ConfigSet{
		Namespace: namespace,
		Registry:  registry,
		Config:    config,
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
