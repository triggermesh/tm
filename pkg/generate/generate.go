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

package generate

import (
	"fmt"
	"strings"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"gopkg.in/yaml.v2"
)

// Project structure contains generic fields to generate sample knative service
type Project struct {
	Name      string
	Namespace string
	Runtime   string
}

// Generate accept Project object and creates tm-deployable project structure
// with all required manifests, functions, etc
func (p *Project) Generate(clientset *client.ConfigSet) error {
	p.Runtime = strings.TrimLeft(p.Runtime, "-")
	samples := NewTable()
	if p.Runtime == "" || p.Runtime == "h" || p.Runtime == "help" {
		return p.help(samples)
	}
	sample, exist := (*samples)[p.Runtime]
	if !exist {
		return p.help(samples)
	}

	var buildArgs []string
	if sample.handler != "" {
		buildArgs = append(buildArgs, fmt.Sprintf("HANDLER=%s", sample.handler))
	}

	provider := file.TriggermeshProvider{
		Name:     "triggermesh",
		Registry: client.Registry,
	}

	functionName := fmt.Sprintf("%s-function", p.Runtime)
	functions := map[string]file.Function{
		functionName: {
			Source:    sample.source,
			Runtime:   sample.runtime,
			Buildargs: buildArgs,
			Environment: map[string]string{
				"foo": "bar",
			},
		},
	}

	if sample.apiGateway {
		functions[functionName].Environment["EVENT"] = "API_GATEWAY"
	}

	template := file.Definition{
		Service:     "demo-service",
		Description: "Sample knative service",
		Provider:    provider,
		Functions:   functions,
	}

	manifest, err := yaml.Marshal(&template)
	if err != nil {
		return err
	}
	if client.Dry {
		fmt.Printf("%s/%s:\n---\n%s\n\n", p.Runtime, sample.source, sample.function)
		fmt.Printf("%s/%s:\n---\n%s\n", p.Runtime, manifestName, manifest)
		return nil
	}

	if err := file.MakeDir(p.Runtime); err != nil {
		return err
	}

	for _, dep := range sample.dependencies {
		if err := file.Write(fmt.Sprintf("%s/%s", p.Runtime, dep.name), dep.data); err != nil {
			return fmt.Errorf("writing dependencies to file: %s", err)
		}
	}
	if err := file.Write(fmt.Sprintf("%s/%s", p.Runtime, sample.source), sample.function); err != nil {
		return fmt.Errorf("writing function to file: %s", err)
	}
	if err := file.Write(fmt.Sprintf("%s/%s", p.Runtime, manifestName), string(manifest)); err != nil {
		return fmt.Errorf("writing manifest to file: %s", err)
	}
	fmt.Printf("Sample %s project has been created\n", p.Runtime)
	fmt.Printf("%s/%s\t\t- function code\n%s/%s\t\t- service manifest\n", p.Runtime, sample.source, p.Runtime, manifestName)
	fmt.Printf("You can deploy this project using \"tm deploy -f %s --wait\" command\n", p.Runtime)
	return nil
}

func (p *Project) help(samples *SamplesTable) error {
	fmt.Printf("Please specify one of available runtimes (e.g. \"tm generate --go\"):\n")
	for runtime := range *samples {
		fmt.Printf("--%s\n", runtime)
	}
	return fmt.Errorf("runtime not found")
}
