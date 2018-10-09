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

package deploy

import (
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

const (
	tmpPath = "/tmp"
)

var (
	df = "Dockerfile"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy knative resource",
}

func NewDeployCmd(clientset *client.ClientSet) *cobra.Command {
	deployCmd.AddCommand(cmdDeployService(clientset))
	deployCmd.AddCommand(cmdDeployBuildTemplate(clientset))
	return deployCmd
}
