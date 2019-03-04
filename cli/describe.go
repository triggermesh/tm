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

package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var (
	output string
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Details of knative resources",
}

func newDescribeCmd(clientset *client.ConfigSet) *cobra.Command {
	describeCmd.AddCommand(cmdDescribeBuild(clientset))
	describeCmd.AddCommand(cmdDescribeBuildTemplate(clientset))
	// describeCmd.AddCommand(cmdDescribeConfiguration(clientset))
	// describeCmd.AddCommand(cmdDescribeRevision(clientset))
	// describeCmd.AddCommand(cmdDescribeRoute(clientset))
	describeCmd.AddCommand(cmdDescribeService(clientset))
	describeCmd.AddCommand(cmdDescribeChannel(clientset))
	return describeCmd
}

func cmdDescribeBuild(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Args:    cobra.ExactArgs(1),
		Short:   "Build details",
		Run: func(cmd *cobra.Command, args []string) {
			b.Name = args[0]
			b.Namespace = client.Namespace
			output, err := b.Describe(clientset)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(output))
		},
	}
}

func cmdDescribeBuildTemplate(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates"},
		Args:    cobra.ExactArgs(1),
		Short:   "BuildTemplate details",
		Run: func(cmd *cobra.Command, args []string) {
			bt.Name = args[0]
			bt.Namespace = client.Namespace
			output, err := bt.Describe(clientset)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(output))
		},
	}
}

func cmdDescribeChannel(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "channel",
		Aliases: []string{"channels"},
		Args:    cobra.ExactArgs(1),
		Short:   "Channel details",
		Run: func(cmd *cobra.Command, args []string) {
			c.Name = args[0]
			c.Namespace = client.Namespace
			output, err := c.Describe(clientset)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(output))
		},
	}
}

func cmdDescribeService(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "service",
		Aliases: []string{"services"},
		Args:    cobra.ExactArgs(1),
		Short:   "Knative service configuration details",
		Run: func(cmd *cobra.Command, args []string) {
			s.Name = args[0]
			s.Namespace = client.Namespace
			output, err := s.Describe(clientset)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(output))
		},
	}
}
