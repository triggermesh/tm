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

	eventSourcesApi "github.com/knative/eventing-sources/pkg/client/clientset/versioned"
	eventingApi "github.com/knative/eventing/pkg/client/clientset/versioned"
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

var (
	Namespace string
	Registry  string
	Output    string
	Debug     bool
	Dry       bool
	Wait      bool
)

// ConfigSet contains different information that may be needed by underlying functions
type ConfigSet struct {
	Core         *kubernetes.Clientset
	Build        *buildApi.Clientset
	Serving      *servingApi.Clientset
	Eventing     *eventingApi.Clientset
	EventSources *eventSourcesApi.Clientset

	Config *rest.Config
}

type config struct {
	Contexts []struct {
		Name    string
		Context struct {
			Namespace string
		}
	}
	CurrentContext string `json:"current-context"`
}

func getNamespace(kubeCfgFile string) string {
	namespace := "default"
	data, err := ioutil.ReadFile(kubeCfgFile)
	if err != nil {
		log.Printf("Can't read config file: %s\n", err)
		return namespace
	}
	var c config
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Printf("Can't parse config body: %s\n", err)
		return namespace
	}
	for _, context := range c.Contexts {
		if context.Name == c.CurrentContext {
			if context.Context.Namespace != "" {
				namespace = context.Context.Namespace
			}
			break
		}
	}
	return namespace
}

func getInClusterNamespace() string {
	data, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "default"
	}
	return string(data)
}

func ConfigPath(cfgFile string) string {
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
	} else if _, err := os.Stat(homeDir + confPath); err == nil {
		cfgFile = homeDir + confPath
	} else if _, err := os.Stat(kubeconfig); err == nil {
		cfgFile = kubeconfig
	} else {
		cfgFile = homeDir + "/.kube/config"
	}
	return cfgFile
}

// NewClient returns ConfigSet created from available configuration file or from in-cluster environment
func NewClient(cfgFile string) (ConfigSet, error) {
	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		log.Printf("%s, falling back to in-cluster configuration\n", err)
		if config, err = rest.InClusterConfig(); err != nil {
			log.Fatalln("Can't read config file")
		}
		if len(Namespace) == 0 {
			Namespace = getInClusterNamespace()
		}
	} else if len(Namespace) == 0 {
		Namespace = getNamespace(cfgFile)
	}

	c := ConfigSet{
		Config: config,
	}

	if c.Eventing, err = eventingApi.NewForConfig(config); err != nil {
		return c, err
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
	if c.EventSources, err = eventSourcesApi.NewForConfig(config); err != nil {
		return c, err
	}

	return c, nil
}
