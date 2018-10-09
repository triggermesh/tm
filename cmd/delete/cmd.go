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

package delete

import (
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete knative resource",
}

func NewDeleteCmd(clientset *client.ClientSet) *cobra.Command {
	deleteCmd.AddCommand(cmdDeleteConfiguration(clientset))
	deleteCmd.AddCommand(cmdDeleteBuildTemplate(clientset))
	deleteCmd.AddCommand(cmdDeleteRevision(clientset))
	deleteCmd.AddCommand(cmdDeleteService(clientset))
	deleteCmd.AddCommand(cmdDeleteBuild(clientset))
	deleteCmd.AddCommand(cmdDeleteRoute(clientset))
	return deleteCmd
}
