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
func (s *Service) DeployYAML(clientset *client.ConfigSet) (services []Service, err error) {
	var root bool
	if file.IsGit(s.YAML) {
		fmt.Printf("Cloning %s\n", s.YAML)
		path, err := file.Clone(s.YAML)
		if err != nil {
			return nil, err
		}
		s.YAML = path + "/serverless.yaml"
	}
	if !file.IsLocal(s.YAML) {
		return nil, fmt.Errorf("Can't read %s", s.YAML)
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
		// We are in the root service
		root = true
		s.Name = definition.Service
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
	services = s.newServices(definition, workdir)

	for _, service := range services {
		if err := service.DeployService(clientset); err != nil {
			return nil, err
		}
	}

	for _, include := range definition.Include {
		s.YAML = workdir + "/" + include
		if file.IsRemote(include) {
			s.YAML = include
		}
		subServices, err := s.DeployYAML(clientset)
		if err != nil {
			return nil, err
		}
		services = append(services, subServices...)
	}

	if root {
		if err = s.removeOrphans(services, clientset); err != nil {
			return nil, err
		}
	}
	return services, nil
}

func (s *Service) newServices(definition file.YAML, path string) []Service {
	var services []Service
	for name, function := range definition.Functions {
		var service Service

		service.Source = function.Handler
		if path != "." && path != "./." && !file.IsRemote(service.Source) {
			service.Source = fmt.Sprintf("%s/%s", path, service.Source)
		}
		service.Wait = s.Wait
		service.ResultImageTag = "latest"
		service.Labels = function.Labels
		service.Labels = append(service.Labels, "service:"+s.Name)
		if len(definition.Service) != 0 && s.Name != definition.Service {
			name = fmt.Sprintf("%s-%s", definition.Service, name)
		}
		service.Name = fmt.Sprintf("%s-%s", s.Name, name)

		if len(function.Runtime) != 0 {
			service.Buildtemplate = function.Runtime
		} else if len(definition.Provider.Runtime) != 0 {
			service.Buildtemplate = definition.Provider.Runtime
		} else if len(s.Buildtemplate) != 0 {
			service.Buildtemplate = s.Buildtemplate
		}
		service.BuildArgs = function.Buildargs
		service.Annotations = make(map[string]string)
		if len(function.Description) != 0 {
			service.Annotations["Description"] = function.Description
		} else if len(definition.Description) != 0 {
			service.Annotations["Description"] = definition.Description
		}
		for k, v := range definition.Provider.Environment {
			service.Env = append(service.Env, k+":"+v)
		}
		for k, v := range function.Environment {
			service.Env = append(service.Env, k+":"+v)
		}
		services = append(services, service)
	}
	return services
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
