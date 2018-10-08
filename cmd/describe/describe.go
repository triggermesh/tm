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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var (
	log    *logrus.Logger
	output string
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Details of knative resources",
}

func NewDescribeCmd(clientset *client.ClientSet, log *logrus.Logger, output *string) *cobra.Command {
	describeCmd.AddCommand(cmdDescribeBuild(clientset))
	describeCmd.AddCommand(cmdDescribeBuildtemplate(clientset))
	describeCmd.AddCommand(cmdDescribeConfiguration(clientset))
	describeCmd.AddCommand(cmdDescribeRevision(clientset))
	describeCmd.AddCommand(cmdDescribeRoute(clientset))
	describeCmd.AddCommand(cmdDescribeService(clientset))
	return describeCmd
}
