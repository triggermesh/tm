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
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Output contains input-output writer interface
var Output io.Writer = os.Stdout
var yamlFile = "serverless.yaml"

type status struct {
	Message string
	Error   error
}

// DeployYAML accepts service YAML manifest and deploys it to cluster
func (s *Service) DeployYAML(yamlFile string, functionsToDeploy []string, threads int, clientset *client.ConfigSet) error {
	services, err := s.ManifestToServices(yamlFile)
	if err != nil {
		return err
	}

	var functions []Service
	for _, service := range services {
		if s.inList(service.Name, functionsToDeploy) {
			functions = append(functions, service)
		}
	}

	removeOrphans := (len(functionsToDeploy) == 0)

	return s.DeployFunctions(functions, removeOrphans, threads, clientset)
}

// DeployFunctions creates a deployment worker pool, reads provided Service array and
// if service is in list to deploy, sends it to the worker pool with given concurrency rate.
// After deployment it checks which functions from current service are left untouched
// and removes them as orphans
func (s *Service) DeployFunctions(functions []Service, removeOrphans bool, threads int, clientset *client.ConfigSet) error {
	jobs := make(chan Service, 100)
	results := make(chan status, 100)
	defer close(jobs)
	defer close(results)

	for w := 0; w < threads; w++ {
		go deploymentWorker(jobs, results, clientset)
	}

	var inProgress int
	for _, function := range functions {
		jobs <- function
		inProgress++
	}

	var errs bool
	for i := 0; i < inProgress; i++ {
		if r := <-results; r.Error != nil {
			errs = true
			fmt.Fprintln(Output, r.Error)
		} else {
			fmt.Fprintln(Output, r.Message)
		}
	}

	if removeOrphans && !client.Dry {
		if err := s.removeOrphans(functions, clientset); err != nil {
			return err
		}
	}

	if errs {
		return fmt.Errorf("There were errors during manifest deployment")
	}
	return nil
}

// DeleteYAML creates deletion worker pool and removes functions listed in provided YAML manifest
func (s *Service) DeleteYAML(yamlFile string, functionsToDelete []string, threads int, clientset *client.ConfigSet) error {
	jobs := make(chan Service, 100)
	results := make(chan status, 100)
	defer close(jobs)
	defer close(results)

	for w := 0; w < threads; w++ {
		go deletionWorker(jobs, results, clientset)
	}

	functions, err := s.ManifestToServices(yamlFile)
	if err != nil {
		return err
	}

	var inProgress int
	for _, function := range functions {
		if !s.inList(function.Name, functionsToDelete) {
			continue
		}
		clientset.Log.Infof("Deleting %s\n", function.Name)
		jobs <- function
		inProgress++
	}

	for i := 0; i < inProgress; i++ {
		if r := <-results; r.Error != nil {
			fmt.Fprintln(Output, r.Error)
		}
	}
	return nil
}

// ManifestToServices parses and validates YAML manifest and returns an array of Service objects
func (s *Service) ManifestToServices(YAML string) ([]Service, error) {
	var err error
	if YAML, err = getYAML(YAML); err != nil {
		return nil, err
	}
	definition, err := file.ParseManifest(YAML)
	if err != nil {
		return nil, err
	}

	err = definition.Validate()
	if err != nil {
		return nil, err
	}

	s.setupParentVars(definition)

	functions := s.parseFunctions(definition.Functions, path.Dir(YAML))
	includedFunctions, err := s.parseIncludes(definition.Include, path.Dir(YAML))
	if err != nil {
		return nil, err
	}
	return append(functions, includedFunctions...), nil
}

func (s *Service) parseIncludes(includes []string, workdir ...string) ([]Service, error) {
	var services []Service
	var definition file.Definition
	for _, include := range includes {
		if !file.IsRemote(include) && len(workdir) == 1 {
			include = path.Join(workdir[0], include)
		}
		YAML, err := getYAML(include)
		if err != nil {
			return []Service{}, err
		}
		definition, err = file.ParseManifest(YAML)
		if err != nil {
			return []Service{}, err
		}
		services = append(services, s.parseFunctions(definition.Functions, path.Dir(YAML))...)
	}

	return services, nil
}

func (s *Service) parseFunctions(functions map[string]file.Function, workdir ...string) []Service {
	var services []Service
	for name, function := range functions {
		service := s.serviceObject(function)
		service.Name = fmt.Sprintf("%s-%s", s.Name, name)
		service.Labels = append(service.Labels, "service:"+s.Name)
		if !file.IsRemote(service.Source) && len(workdir) == 1 {
			service.Source = path.Join(workdir[0], service.Source)
		}
		service.parseSchedule(function.Events)
		services = append(services, service)
	}
	return services
}

func (s *Service) inList(name string, list []string) bool {
	listed := true
	for _, v := range list {
		listed = false
		if v == name || name == fmt.Sprintf("%s-%s", s.Name, v) {
			return true
		}
	}
	return listed
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
			}
		}
	}
}

func (s *Service) setupParentVars(definition file.Definition) {
	s.Annotations = make(map[string]string)
	s.Name = definition.Service
	s.EnvSecrets = definition.Provider.EnvSecrets
	s.PullPolicy = definition.Provider.PullPolicy
	s.Runtime = definition.Provider.Runtime
	s.BuildTimeout = definition.Provider.Buildtimeout

	if len(s.Namespace) == 0 {
		s.Namespace = definition.Provider.Namespace
	}
	for k, v := range definition.Provider.Annotations {
		s.Annotations[k] = v
	}
	if len(definition.Description) != 0 {
		s.Annotations["Description"] = definition.Description
	}
	for k, v := range definition.Provider.Environment {
		s.Env = append(s.Env, k+":"+v)
	}
}

func (s *Service) serviceObject(function file.Function) Service {
	service := Service{
		Source:         function.Source,
		Revision:       function.Revision,
		Namespace:      s.Namespace,
		Concurrency:    function.Concurrency,
		Runtime:        function.Runtime,
		Labels:         function.Labels,
		PullPolicy:     s.PullPolicy,
		ResultImageTag: "latest",
		BuildArgs:      function.Buildargs,
		BuildTimeout:   s.BuildTimeout,
		Env:            s.Env,
		Annotations:    make(map[string]string),
		EnvSecrets:     append(s.EnvSecrets, function.EnvSecrets...),
	}
	// For back-compatibility with old "handler" field
	if len(function.Handler) != 0 {
		service.Source = function.Handler
	}
	for k, v := range function.Environment {
		service.Env = append(service.Env, k+":"+v)
	}
	for k, v := range s.Annotations {
		service.Annotations[k] = v
	}
	for k, v := range function.Annotations {
		service.Annotations[k] = v
	}
	if len(service.Runtime) == 0 {
		service.Runtime = s.Runtime
	}
	if len(function.Description) != 0 {
		service.Annotations["Description"] = fmt.Sprintf("%s\n%s", service.Annotations["Description"], function.Description)
	}
	return service
}

func (s *Service) removeOrphans(created []Service, clientset *client.ConfigSet) error {
	list, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).List(metav1.ListOptions{
		LabelSelector: "service=" + s.Name,
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
			orphan := Service{
				Name:      existing.Name,
				Namespace: existing.Namespace,
			}
			fmt.Fprintf(Output, "Removing orphaned function %s\n", orphan.Name)
			if err = orphan.Delete(clientset); err != nil {
				return err
			}
		}
	}
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
	if file.IsDir(filepath) {
		filepath = path.Join(filepath, yamlFile)
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

func deploymentWorker(services <-chan Service, results chan<- status, clientset *client.ConfigSet) {
	for service := range services {
		output, err := service.Deploy(clientset)
		results <- status{
			Message: output,
			Error:   err,
		}
	}
}

func deletionWorker(services <-chan Service, results chan<- status, clientset *client.ConfigSet) {
	for service := range services {
		results <- status{
			Error: service.Delete(clientset),
		}
	}
}
