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

package get

import (
	"encoding/json"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	yaml "gopkg.in/yaml.v2"
)

var (
	table  *uitable.Table
	output string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve resources from k8s cluster",
}

// NewGetCmd returns "Get" cobra CLI command with its subcommands
func NewGetCmd(clientset *client.ConfigSet) *cobra.Command {
	getCmd.AddCommand(cmdListBuild(clientset))
	getCmd.AddCommand(cmdListBuildTemplates(clientset))
	getCmd.AddCommand(cmdListConfigurations(clientset))
	getCmd.AddCommand(cmdListRevision(clientset))
	getCmd.AddCommand(cmdListRoute(clientset))
	getCmd.AddCommand(cmdListService(clientset))
	getCmd.AddCommand(cmdListChannels(clientset))

	table = uitable.New()
	table.Wrap = true
	table.MaxColWidth = 50

	return getCmd
}

// Format sets Get command output format
func Format(encode *string) {
	if encode == nil || string(*encode) == "json" {
		output = "json"
	} else if encode != nil && string(*encode) == "yaml" {
		output = "yaml"
	}
}

func encode(data interface{}) (string, error) {
	if output == "yaml" {
		o, err := yaml.Marshal(data)
		return string(o), err
	}
	o, err := json.MarshalIndent(data, "", "    ")
	return string(o), err
}
