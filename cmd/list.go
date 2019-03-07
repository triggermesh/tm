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

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/output"
)

var (
	data string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"list", "describe"},
	Short:   "Retrieve resources from k8s cluster",
}

// NewGetCmd returns "Get" cobra CLI command with its subcommands
func newGetCmd(clientset *client.ConfigSet) *cobra.Command {
	getCmd.AddCommand(cmdListBuild(clientset))
	getCmd.AddCommand(cmdListBuildTemplates(clientset))
	getCmd.AddCommand(cmdListConfigurations(clientset))
	getCmd.AddCommand(cmdListRevision(clientset))
	getCmd.AddCommand(cmdListRoute(clientset))
	getCmd.AddCommand(cmdListService(clientset))
	getCmd.AddCommand(cmdListChannels(clientset))
	return getCmd
}

func cmdListBuild(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "List of knative build resources",
		Run: func(cmd *cobra.Command, args []string) {
			b.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := b.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				b.Name = args[0]
				build, err := b.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(build, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListBuildTemplates(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates"},
		Short:   "List of buildtemplates",
		Run: func(cmd *cobra.Command, args []string) {
			bt.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := bt.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				bt.Name = args[0]
				buildtemplate, err := bt.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(buildtemplate, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListChannels(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "channel",
		Aliases: []string{"channels"},
		Short:   "List of knative channel resources",
		Run: func(cmd *cobra.Command, args []string) {
			c.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := c.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				c.Name = args[0]
				channel, err := c.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(channel, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListService(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "service",
		Aliases: []string{"services"},
		Short:   "List of knative service resources",
		Run: func(cmd *cobra.Command, args []string) {
			s.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := s.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				s.Name = args[0]
				service, err := s.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(service, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListConfigurations(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"configurations"},
		Short:   "List of configurations",
		Run: func(cmd *cobra.Command, args []string) {
			cf.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := cf.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				cf.Name = args[0]
				configuration, err := cf.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(configuration, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListRevision(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "revision",
		Aliases: []string{"revisions"},
		Short:   "List of knative revision resources",
		Run: func(cmd *cobra.Command, args []string) {
			r.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := r.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				r.Name = args[0]
				revision, err := r.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(revision, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListRoute(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "route",
		Aliases: []string{"routes"},
		Short:   "List of knative routes resources",
		Run: func(cmd *cobra.Command, args []string) {
			rt.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := rt.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				rt.Name = args[0]
				route, err := r.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				data, err = output.Describe(route, client.Output)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}
