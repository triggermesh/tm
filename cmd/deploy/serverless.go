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

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
)

// TODO Cleanup and simplify

// YAML deploys functions defined in serverless.yaml file
func (s *Service) fromYAML(clientset *client.ConfigSet) (err error) {
	if !file.IsLocal(s.YAML) {
		if s.YAML, err = file.Download(s.YAML); err != nil {
			return errors.New("Can't get YAML file")
		}
	} else if err != nil {
		return err
	}
	definition, err := file.ParseServerlessYAML(s.YAML)
	if err != nil {
		return err
	}
	if len(definition.Provider.Name) != 0 && definition.Provider.Name != "triggermesh" {
		return fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}

	if len(definition.Service) != 0 {
		s.Name = definition.Service
	}
	if len(definition.Provider.Runtime) != 0 {
		s.Buildtemplate = definition.Provider.Runtime
	}
	workdir := p.Dir(s.YAML)
	services, err := s.newServices(definition, workdir)
	if err != nil {
		return err
	}

	for _, service := range services {
		if err := service.DeployService(clientset); err != nil {
			return err
		}
	}

	for _, include := range definition.Include {
		s.YAML = workdir + "/" + include
		if err := s.fromYAML(clientset); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) newServices(definition file.YAML, path string) ([]Service, error) {
	var services []Service
	for name, function := range definition.Functions {
		var service Service

		service.Source = function.Handler
		if path != "." && path != "./." && !file.IsRemote(service.Source) {
			service.Source = fmt.Sprintf("%s/%s", path, service.Source)
		}
		service.Wait = s.Wait
		service.Name = name
		service.ResultImageTag = "latest"
		service.Labels = function.Labels
		if len(s.Name) != 0 {
			service.Name = fmt.Sprintf("%s-%s", s.Name, service.Name)
		}
		if len(function.Runtime) != 0 {
			service.Buildtemplate = function.Runtime
		} else if len(definition.Provider.Runtime) != 0 {
			service.Buildtemplate = definition.Provider.Runtime
		} else if len(s.Buildtemplate) != 0 {
			service.Buildtemplate = s.Buildtemplate
		}
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
	return services, nil
}
