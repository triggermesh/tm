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
	p "path"

	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO Cleanup and simplify

// DeployYAML deploys functions defined in serverless.yaml file
func (s *Service) DeployYAML(functions []string, clientset *client.ConfigSet) (services []Service, err error) {
	var root bool
	if s.YAML, err = getYAML(s.YAML); err != nil {
		return nil, err
	}
	definition, err := file.ParseServerlessYAML(s.YAML)
	if err != nil {
		return nil, err
	}
	if len(definition.Provider.Name) != 0 && definition.Provider.Name != "triggermesh" {
		return nil, fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}
	if len(s.Name) == 0 && len(definition.Service) == 0 {
		return nil, errors.New("Service name can't be empty")
	}
	if len(s.Name) == 0 {
		root = true
		s.Name = definition.Service
		s.RegistrySecret = definition.Provider.RegistrySecret
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
	workdir := p.Dir(s.YAML)

	for name, function := range definition.Functions {
		pass := false
		for _, v := range functions {
			if v == name {
				pass = true
				break
			}
		}
		if len(functions) != 0 && !pass {
			continue
		}
		service := newService(function)
		service.Wait = s.Wait
		service.RegistrySecret = s.RegistrySecret
		service.Name = name
		if len(definition.Service) != 0 {
			service.Name = fmt.Sprintf("%s-%s", definition.Service, service.Name)
		}
		if s.Name != definition.Service {
			service.Name = fmt.Sprintf("%s-%s", s.Name, service.Name)
		}
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
			service.Annotations["Description"] = fmt.Sprintf("%s\n%s", definition.Description, service.Annotations["Description"])
		}
		if len(s.Annotations["Description"]) != 0 {
			service.Annotations["Description"] = fmt.Sprintf("%s\n%s", s.Annotations["Description"], service.Annotations["Description"])
		}
		for k, v := range definition.Provider.Environment {
			service.Env = append(service.Env, k+":"+v)
		}
		service.Env = append(service.Env, s.Env...)

		if err := service.Deploy(clientset); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	for _, include := range definition.Include {
		s.YAML = workdir + "/" + include
		if file.IsRemote(include) {
			s.YAML = include
		}
		subServices, err := s.DeployYAML(functions, clientset)
		if err != nil {
			return nil, err
		}
		services = append(services, subServices...)
	}

	if root && len(functions) == 0 {
		if err = s.removeOrphans(services, clientset); err != nil {
			return nil, err
		}
	}
	return services, nil
}

func getYAML(path string) (string, error) {
	if file.IsGit(path) {
		// fmt.Printf("Cloning %s\n", path)
		localPath, err := file.Clone(path)
		if err != nil {
			return "", err
		}
		path = localPath + "/serverless.yaml"
	}
	if !file.IsLocal(path) {
		return "", fmt.Errorf("Can't read %s", service.YAML)
	}
	return path, nil
}

func newService(function file.Function) Service {
	service := Service{
		Source:         function.Handler,
		Buildtemplate:  function.Runtime,
		Labels:         append(function.Labels, "service:"+service.Name),
		ResultImageTag: "latest",
		BuildArgs:      function.Buildargs,
		Annotations:    make(map[string]string),
	}
	service.Annotations["Description"] = function.Description
	for k, v := range function.Environment {
		service.Env = append(service.Env, k+":"+v)
	}
	return service
}

func (s *Service) removeOrphans(services []Service, clientset *client.ConfigSet) error {
	list, err := clientset.Serving.ServingV1alpha1().Revisions(clientset.Namespace).List(metav1.ListOptions{
		IncludeUninitialized: true,
		LabelSelector:        "service=" + s.Name,
	})
	if err != nil {
		return err
	}
	for _, oldService := range list.Items {
		if len(oldService.OwnerReferences) != 1 {
			continue
		}
		orphan := true
		for _, newService := range services {
			if newService.Name == oldService.OwnerReferences[0].Name {
				orphan = false
				break
			}
		}
		if orphan {
			s := delete.Service{
				Name: oldService.OwnerReferences[0].Name,
			}
			if err = s.DeleteService(clientset); err == nil {
				fmt.Printf("Removing orphaned service: %s\n", s.Name)
			}
		}
	}

	return nil
}
