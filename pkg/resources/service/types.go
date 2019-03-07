// Copyright 2019 TriggerMesh, Inc
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

import "encoding/json"

// Service represents knative service structure
type Service struct {
	Name           string
	Namespace      string
	Registry       string
	Source         string
	Revision       string
	PullPolicy     string
	Concurrency    int
	ResultImageTag string
	Buildtemplate  string
	RegistrySecret string // Does not belong to the service, need to be deleted
	Env            []string
	EnvSecrets     []string
	Annotations    map[string]string
	Labels         []string
	BuildArgs      []string
	BuildTimeout   string
	Cronjob        struct {
		Schedule string
		Data     string
	}
}

type registryAuths struct {
	Auths registry
}

type credentials struct {
	Username string
	Password string
}

type registry map[string]credentials

func encode(data interface{}) ([]byte, error) {
	// if output == "yaml" {
	// return yaml.Marshal(data)
	// }
	return json.MarshalIndent(data, "", "    ")
}
