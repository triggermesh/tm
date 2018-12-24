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

package deploy

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO Cleanup and simplify

// DeployYAML deploys functions defined in serverless.yaml file
func (s *Service) DeployYAML(YAML string, functionsToDeploy []string, clientset *client.ConfigSet) (services []Service, err error) {
	if YAML, err = getYAML(YAML); err != nil {
		return nil, err
	}
	definition, err := file.ParseServerlessYAML(YAML)
	if err != nil {
		return nil, err
	}
	if len(definition.Provider.Name) != 0 && definition.Provider.Name != "triggermesh" {
		return nil, fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}
	if len(definition.Service) == 0 {
		return nil, errors.New("Service name can't be empty")
	}
	if len(s.Name) == 0 {
		s.setupParentVars(definition)
	}
	if len(definition.Provider.Registry) != 0 {
		clientset.Registry = definition.Provider.Registry
	}
	if len(definition.Provider.Namespace) != 0 {
		clientset.Namespace = definition.Provider.Namespace
	}
	if len(definition.Provider.Runtime) != 0 {
		s.Buildtemplate = definition.Provider.Runtime
	}
	prefix := s.Name
	if s.Name != definition.Service {
		prefix = fmt.Sprintf("%s-%s", prefix, definition.Service)
	}
	workdir := path.Dir(YAML)

	for name, function := range definition.Functions {
		if !inList(name, functionsToDeploy) {
			continue
		}
		service := s.serviceObject(function)
		service.Labels = append(service.Labels, "service:"+prefix)
		service.Name = fmt.Sprintf("%s-%s", prefix, name)
		if workdir != "." && workdir != "./." && !file.IsRemote(service.Source) {
			service.Source = fmt.Sprintf("%s/%s", workdir, service.Source)
		}
		if len(service.Buildtemplate) == 0 {
			service.Buildtemplate = definition.Provider.Runtime
		}
		if len(service.Buildtemplate) == 0 {
			service.Buildtemplate = s.Buildtemplate
		}
		if len(definition.Description) != 0 {
			service.Annotations["Description"] = fmt.Sprintf("%s\n%s", service.Annotations["Description"], definition.Description)
		}
		if len(function.Description) != 0 {
			service.Annotations["Description"] = fmt.Sprintf("%s\n%s", service.Annotations["Description"], function.Description)
		}

		if err := service.Deploy(clientset); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if len(functionsToDeploy) == 0 {
		if err = removeOrphans(services, prefix, clientset); err != nil {
			return nil, err
		}
	}

	for _, include := range definition.Include {
		YAML = workdir + "/" + include
		if file.IsRemote(include) {
			YAML = include
		}
		if _, err := s.DeployYAML(YAML, functionsToDeploy, clientset); err != nil {
			return nil, err
		}
	}

	return services, nil
}

func inList(name string, functionsToDeploy []string) bool {
	deployThis := true
	if len(functionsToDeploy) != 0 {
		deployThis = false
		for _, v := range functionsToDeploy {
			if v == name {
				return true
			}
		}
	}
	return deployThis
}

func getYAML(filepath string) (string, error) {
	if file.IsGit(filepath) {
		localfilepath, err := file.Clone(filepath)
		if err != nil {
			return "", err
		}
		filepath = localfilepath + "/serverless.yaml"
	}
	if !file.IsLocal(filepath) {
		/* Add a secondary check against /serverless.yml */
		filepath = strings.TrimSuffix(filepath, ".yaml")
		filepath = filepath + ".yml"

		if !file.IsLocal(filepath) {
			return "", fmt.Errorf("Can't read %s", filepath)
		}
	}
	return filepath, nil
}

func (s *Service) setupParentVars(definition file.Definition) {
	s.Name = definition.Service
	s.RegistrySecret = definition.Provider.RegistrySecret
	if len(definition.Description) != 0 {
		s.Annotations = map[string]string{
			"Description": definition.Description,
		}
	}
	// workaround to get rid of double description
	definition.Description = ""
	for k, v := range definition.Provider.Environment {
		s.Env = append(s.Env, k+":"+v)
	}
}

func (s *Service) serviceObject(function file.Function) Service {
	service := Service{
		Source:         function.Handler,
		Buildtemplate:  function.Runtime,
		Labels:         function.Labels,
		ResultImageTag: "latest",
		BuildArgs:      function.Buildargs,
		Wait:           s.Wait,
		RegistrySecret: s.RegistrySecret,
		Annotations: map[string]string{
			"Description": s.Annotations["Description"],
		},
	}
	service.Env = s.Env
	for k, v := range function.Environment {
		service.Env = append(service.Env, k+":"+v)
	}
	return service
}

func removeOrphans(created []Service, label string, clientset *client.ConfigSet) error {
	list, err := clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).List(metav1.ListOptions{
		IncludeUninitialized: true,
		LabelSelector:        "service=" + label,
	})
	if err != nil {
		return err
	}

	for _, existing := range list.Items {
		orphaned := true
		for _, newService := range created {
			if newService.Name == existing.Name {
				orphaned = false
				break
			}
		}
		if orphaned {
			orphan := delete.Service{
				Name: existing.Name,
			}
			if err = orphan.DeleteService(clientset); err == nil {
				fmt.Printf("Removing orphaned service: %s\n", existing.Name)
			}
		}
	}

	return nil
}
