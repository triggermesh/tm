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

package service

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	yaml "gopkg.in/yaml.v2"
)

// TODO Cleanup and simplify

var yamlFile = "serverless.yaml"

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
		s.Registry = definition.Provider.Registry
	}
	if len(definition.Provider.Namespace) != 0 {
		s.Namespace = definition.Provider.Namespace
	}
	if len(definition.Provider.Runtime) != 0 {
		s.Buildtemplate = definition.Provider.Runtime
	}
	prefix := s.Name
	if s.Name != definition.Service {
		prefix = fmt.Sprintf("%s-%s", prefix, definition.Service)
	}
	workdir := path.Dir(YAML)

	var wg sync.WaitGroup
	for name, function := range definition.Functions {
		if !inList(name, functionsToDeploy) {
			continue
		}

		service := s.serviceObject(function)
		if len(function.Handler) != 0 {
			fmt.Printf("Warning! Please change \"handler:%s\" to \"source:%s\" for function \"%s\" in serverless.yaml. Parameter \"handler\" will be deprecated soon\n",
				function.Handler, function.Handler, name)
			service.Source = function.Handler
		}
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

		if len(function.Description) != 0 {
			service.Annotations["Description"] = fmt.Sprintf("%s\n%s", service.Annotations["Description"], function.Description)
		}

		service.parseSchedule(function.Events)

		wg.Add(1)
		go func(service Service) {
			defer wg.Done()
			output, err := service.Deploy(clientset)
			if err != nil {
				fmt.Printf("%s: %s\n", service.Name, err)
			} else {
				fmt.Print(output)
			}
		}(service)
		services = append(services, service)
	}

	if len(functionsToDeploy) == 0 {
		if err = s.removeOrphans(services, prefix, clientset); err != nil {
			return nil, err
		}
	}

	for _, include := range definition.Include {
		YAML = workdir + "/" + include
		if file.IsRemote(include) {
			YAML = include
		}
		wg.Add(1)
		go func(YAML string, functionsToDeploy []string) {
			defer wg.Done()
			if _, err := s.DeployYAML(YAML, functionsToDeploy, clientset); err != nil {
				fmt.Printf("%s: %s\n", YAML, err)
			}
		}(YAML, functionsToDeploy)
	}
	wg.Wait()

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

func (s *Service) parseSchedule(events []map[string]interface{}) {
	for _, v := range events {
		for eventType, event := range v {
			eventBody, err := yaml.Marshal(event)
			if err != nil {
				continue
			}
			switch eventType {
			case "schedule":
				var cron file.Schedule
				if err := yaml.Unmarshal(eventBody, &cron); err != nil {
					continue
				}
				s.Cronjob.Schedule = cron.Rate
				s.Cronjob.Data = cron.Data
			}
		}
	}
}

func (s *Service) setupParentVars(definition file.Definition) {
	s.Name = definition.Service
	s.RegistrySecret = definition.Provider.RegistrySecret
	s.Annotations = make(map[string]string)
	for k, v := range definition.Provider.Annotations {
		s.Annotations[k] = v
	}
	if len(definition.Description) != 0 {
		s.Annotations["Description"] = definition.Description
	}
	for k, v := range definition.Provider.Environment {
		s.Env = append(s.Env, k+":"+v)
	}
	s.EnvSecrets = definition.Provider.EnvSecrets
	s.PullPolicy = definition.Provider.PullPolicy
	s.BuildTimeout = definition.Provider.Buildtimeout
}

func (s *Service) serviceObject(function file.Function) Service {
	service := Service{
		Source:         function.Source,
		Registry:       s.Registry,
		Namespace:      s.Namespace,
		Concurrency:    function.Concurrency,
		Buildtemplate:  function.Runtime,
		Labels:         function.Labels,
		PullPolicy:     s.PullPolicy,
		ResultImageTag: "latest",
		BuildArgs:      function.Buildargs,
		BuildTimeout:   s.BuildTimeout,
		RegistrySecret: s.RegistrySecret,
		Env:            s.Env,
		EnvSecrets:     append(s.EnvSecrets, function.EnvSecrets...),
	}
	service.Annotations = make(map[string]string)
	for k, v := range function.Environment {
		service.Env = append(service.Env, k+":"+v)
	}
	for k, v := range s.Annotations {
		service.Annotations[k] = v
	}
	for k, v := range function.Annotations {
		service.Annotations[k] = v
	}
	return service
}

func (s *Service) removeOrphans(created []Service, label string, clientset *client.ConfigSet) error {
	// list, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).List(metav1.ListOptions{
	// 	IncludeUninitialized: true,
	// 	LabelSelector:        "service=" + label,
	// })
	// if err != nil {
	// 	return err
	// }

	// for _, existing := range list.Items {
	// orphaned := true
	// for _, newService := range created {
	// 	if newService.Name == existing.Name {
	// 		orphaned = false
	// 		break
	// 	}
	// }
	// if orphaned {
	// 	orphan := delete.Service{
	// 		Name: existing.Name,
	// 	}
	// 	if err = orphan.DeleteService(clientset); err == nil {
	// 		fmt.Printf("Removing orphaned service: %s\n", existing.Name)
	// 	}
	// }
	// }

	return nil
}

func (s *Service) DeleteYAML(filepath string, functions []string, clientset *client.ConfigSet) (err error) {
	var wg sync.WaitGroup
	if filepath, err = getYAML(filepath); err != nil {
		return err
	}
	definition, err := file.ParseServerlessYAML(filepath)
	if err != nil {
		return err
	}
	if len(definition.Provider.Name) != 0 && definition.Provider.Name != "triggermesh" {
		return fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}
	if len(s.Name) == 0 && len(definition.Service) == 0 {
		return errors.New("Service name can't be empty")
	}
	if len(s.Name) == 0 {
		s.Name = definition.Service
	}

	for name := range definition.Functions {
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
		if len(definition.Service) != 0 && s.Name != definition.Service {
			name = fmt.Sprintf("%s-%s", definition.Service, name)
		}
		service := Service{
			Name: fmt.Sprintf("%s-%s", s.Name, name),
		}
		service.Namespace = s.Namespace

		wg.Add(1)
		fmt.Printf("Deleting %s\n", service.Name)
		go func(service Service) {
			defer wg.Done()
			if err = service.Delete(clientset); err != nil {
				fmt.Printf("%s: %s\n", service.Name, err)
			}
		}(service)
	}
	for _, include := range definition.Include {
		filepath = path.Dir(filepath) + "/" + include
		if file.IsRemote(include) {
			if filepath, err = file.Clone(include); err != nil {
				return err
			}
			filepath = filepath + "/serverless.yaml"
		}
		wg.Add(1)
		go func(filepath string, functions []string) {
			defer wg.Done()
			if err = s.DeleteYAML(filepath, functions, clientset); err != nil {
				fmt.Printf("%s: %s\n", filepath, err)
			}
		}(filepath, functions)
	}
	wg.Wait()
	return nil
}

func getYAML(filepath string) (string, error) {
	if repository, pathToFile := file.IsGitFile(filepath); len(repository) != 0 {
		filepath = repository
		yamlFile = pathToFile
	}
	if file.IsGit(filepath) {
		localfilepath, err := file.Clone(filepath)
		if err != nil {
			return "", err
		}
		filepath = path.Join(localfilepath, yamlFile)
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
