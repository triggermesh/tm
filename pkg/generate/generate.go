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

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"gopkg.in/yaml.v2"
)

type Project struct {
	Name      string
	Namespace string
	Runtime   string
}

func (p *Project) Generate(clientset *client.ConfigSet) error {
	ss := NewTable()

	sample, exist := (*ss)[p.Runtime]
	if !exist {
		return fmt.Errorf("runtime %q does not exist", p.Runtime)
	}

	var buildArgs []string
	if sample.handler != "" {
		buildArgs = append(buildArgs, fmt.Sprintf("HANDLER=%s", sample.handler))
	}

	provider := file.TriggermeshProvider{
		Name:     "triggermesh",
		Registry: client.Registry,
	}

	functions := map[string]file.Function{
		fmt.Sprintf("%s-function", p.Runtime): file.Function{
			Source:    sample.source,
			Runtime:   sample.runtime,
			Buildargs: buildArgs,
			Environment: map[string]string{
				"foo": "bar",
			},
		},
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
		fmt.Printf("%s:\n---\n%s\n\n", sample.source, sample.function)
		fmt.Printf("%s:\n---\n%s\n", manifestName, manifest)
		return nil
	}
	if err := file.Write(sample.source, sample.function); err != nil {
		return fmt.Errorf("writing function to file: %s", err)
	}
	if err := file.Write(manifestName, string(manifest)); err != nil {
		return fmt.Errorf("writing manifest to file: %s", err)
	}
	fmt.Printf("Sample %s project has been created in current directory\n", p.Runtime)
	fmt.Printf("%s\t\t- function code\n%s\t\t- service manifest\n", sample.source, manifestName)
	fmt.Printf("You can deploy this project using \"tm deploy\" command\n")
	return nil
}
