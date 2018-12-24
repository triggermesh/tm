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

package delete

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
)

// Service structure has minimal required set of fields to delete service
type Service struct {
	Name string
}

// DeleteYAML removes functions defined in serverless.yaml file
func (s *Service) DeleteYAML(filepath string, functions []string, clientset *client.ConfigSet) (err error) {
	var wg sync.WaitGroup
	if file.IsGit(filepath) {
		// fmt.Printf("Cloning %s\n", path)
		if filepath, err = file.Clone(filepath); err != nil {
			return err
		}
		filepath = filepath + "/serverless.yaml"
	}
	if !file.IsLocal(filepath) {
		/* Add a secondary check against serverless.yml */
		filepath = strings.TrimSuffix(filepath, ".yaml")
		filepath = filepath + ".yml"

		if !file.IsLocal(filepath) {
			return errors.New("Can't get YAML file")
		}
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

		wg.Add(1)
		fmt.Printf("Deleting %s\n", service.Name)
		go func(service Service) {
			defer wg.Done()
			if err = service.DeleteService(clientset); err != nil {
				fmt.Println(err)
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
				fmt.Println(err)
			}
		}(filepath, functions)
	}
	wg.Wait()
	return nil
}
