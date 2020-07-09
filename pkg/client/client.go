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
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	buildApi "github.com/knative/build/pkg/client/clientset/versioned"
	tektonTask "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	tektonResource "github.com/tektoncd/pipeline/pkg/client/resource/clientset/versioned"
	triggersApi "github.com/tektoncd/triggers/pkg/client/clientset/versioned"
	logwrapper "github.com/triggermesh/tm/pkg/log"
	printerwrapper "github.com/triggermesh/tm/pkg/printer"
	githubSource "knative.dev/eventing-contrib/github/pkg/client/clientset/versioned"
	eventingApi "knative.dev/eventing/pkg/client/clientset/versioned"
	servingApi "knative.dev/serving/pkg/client/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// gcp package is required for kubectl configs with GCP auth providers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	confPath        = "/.tm/config.json"
	defaultRegistry = "knative.registry.svc.cluster.local"
)

// CLI global flags
var (
	// Namespace to work in passed with "-n" argument or defined in kube configs
	Namespace string
	// Output format for k8s objects in "tm get" result. Can be either "yaml" (default) or "json"
	Output string
	// Debug enables verbose output for CLI commands
	Debug bool
	// Dry run of some commands
	Dry bool
	// Wait till deployment operation finishes
	Wait bool
)

// Registry to store docker images for user services
type Registry struct {
	Host    string
	Secret  string
	SkipTLS bool
}

// ConfigSet contains different information that may be needed by underlying functions
type ConfigSet struct {
	Core            *kubernetes.Clientset
	Build           *buildApi.Clientset
	Serving         *servingApi.Clientset
	Eventing        *eventingApi.Clientset
	GithubSource    *githubSource.Clientset
	TektonPipelines *tektonResource.Clientset
	TektonTasks     *tektonTask.Clientset
	TektonTriggers  *triggersApi.Clientset
	Registry        *Registry
	Log             *logwrapper.StandardLogger
	Printer         *printerwrapper.Printer
	Config          *rest.Config
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

// ConfigPath calculates local path to get tm config from
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
func NewClient(cfgFile string, output ...io.Writer) (ConfigSet, error) {
	var c ConfigSet

	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		log.Printf("%s, falling back to in-cluster configuration\n", err)
		if config, err = rest.InClusterConfig(); err != nil {
			return c, err
		}
		if len(Namespace) == 0 {
			Namespace = getInClusterNamespace()
		}
	} else if len(Namespace) == 0 {
		Namespace = getNamespace(cfgFile)
	}
	c.Config = config
	c.Log = logwrapper.NewLogger()
	if len(output) == 1 {
		c.Printer = printerwrapper.NewPrinter(output[0])
	}
	c.Registry = &Registry{
		Host: defaultRegistry,
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
	if c.TektonPipelines, err = tektonResource.NewForConfig(config); err != nil {
		return c, err
	}
	if c.TektonTasks, err = tektonTask.NewForConfig(config); err != nil {
		return c, err
	}
	if c.TektonTriggers, err = triggersApi.NewForConfig(config); err != nil {
		return c, err
	}
	if c.GithubSource, err = githubSource.NewForConfig(config); err != nil {
		return c, err
	}
	return c, nil
}
