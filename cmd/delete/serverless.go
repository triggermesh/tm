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
	p "path"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
)

// YAML removes functions defined in serverless.yaml file
func YAML(path string, clientset *client.ConfigSet) (err error) {
	if !file.Local(path) {
		if path, err = file.Download(path); err != nil {
			return errors.New("Can't get YAML file")
		}
	}
	definition, err := file.ParseServerlessYAML(path)
	if err != nil {
		return err
	}
	if definition.Provider.Name != "triggermesh" {
		return fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}
	rootServiceName := definition.Service
	for name := range definition.Functions {
		serviceName := name
		if len(rootServiceName) != 0 {
			serviceName = fmt.Sprintf("%s-%s", rootServiceName, serviceName)
		}
		fmt.Printf("Deleting %s\n", serviceName)
		if err = Service(serviceName, clientset); err != nil {
			fmt.Println(err)
		}
	}
	for _, include := range definition.Include {
		definition, err := file.ParseServerlessYAML(p.Dir(path) + "/" + include)
		if err != nil {
			return err
		}
		for serviceName := range definition.Functions {
			if len(rootServiceName) != 0 {
				serviceName = fmt.Sprintf("%s-%s", rootServiceName, serviceName)
			}
			fmt.Printf("Deleting %s\n", serviceName)
			if err = Service(serviceName, clientset); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
