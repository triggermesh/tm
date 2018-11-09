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
	"net/http"
	"os"
	"strings"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/serverless"
)

func fromYAML(path string, clientset *client.ConfigSet) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if path, err = downloadFile(path); err != nil {
			return errors.New("Can't get YAML file")
		}
	} else {
		return err
	}
	definition, err := serverless.Parse(path)
	if err != nil {
		return err
	}
	if definition.Provider.Name != "triggermesh" {
		return fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}
	services, err := newServices(definition)
	if err != nil {
		return err
	}
	for _, service := range services {
		if err := service.DeployService(clientset); err != nil {
			return err
		}
	}
	return nil
}

func newServices(definition serverless.File) ([]Service, error) {
	var services []Service
	for name, function := range definition.Functions {
		var service Service

		service.Source = function.Handler
		service.Wait = s.Wait
		service.Name = name
		if len(definition.Service) != 0 {
			service.Name = fmt.Sprintf("%s-%s", definition.Service, service.Name)
		}
		service.Buildtemplate = definition.Provider.Runtime
		if len(function.Runtime) != 0 {
			service.Buildtemplate = function.Runtime
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

func isLocal(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func isGit(path string) bool {
	if strings.HasSuffix(path, ".git") {
		return true
	}
	if resp, err := http.Get(path); err == nil {
		if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 401 {
			return true
		}
	}
	return false
}

func isRegistry(path string) bool {
	if resp, err := http.Get(path); err == nil {
		if resp.StatusCode == 405 {
			return true
		}
	}
	return false
}
