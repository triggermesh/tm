/*
Copyright (c) 2018 TriggerMesh, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package describe

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	yaml "gopkg.in/yaml.v2"
)

var (
	output string
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Details of knative resources",
}

func NewDescribeCmd(clientset *client.ClientSet) *cobra.Command {
	describeCmd.AddCommand(cmdDescribeBuild(clientset))
	describeCmd.AddCommand(cmdDescribeBuildTemplate(clientset))
	describeCmd.AddCommand(cmdDescribeConfiguration(clientset))
	describeCmd.AddCommand(cmdDescribeRevision(clientset))
	describeCmd.AddCommand(cmdDescribeRoute(clientset))
	describeCmd.AddCommand(cmdDescribeService(clientset))
	return describeCmd
}

func Format(encode *string) {
	if encode == nil || string(*encode) == "json" {
		output = "json"
	} else if encode != nil && string(*encode) == "yaml" {
		output = "yaml"
	}
}

func encode(data interface{}) ([]byte, error) {
	if output == "yaml" {
		return yaml.Marshal(data)
	}
	return json.MarshalIndent(data, "", "    ")
}
