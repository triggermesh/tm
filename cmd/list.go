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
)

// var (
// table *uitable.Table
// )

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"list"},
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

	// table = uitable.New()
	// table.Wrap = true
	// table.MaxColWidth = 50

	return getCmd
}

// // Format sets Get command output format
// func Format(encode *string) {
// 	if encode == nil || string(*encode) == "json" {
// 		output = "json"
// 	} else if encode != nil && string(*encode) == "yaml" {
// 		output = "yaml"
// 	}
// }

// func encode(data interface{}) (string, error) {
// 	if output == "yaml" {
// 		o, err := yaml.Marshal(data)
// 		return string(o), err
// 	}
// 	o, err := json.MarshalIndent(data, "", "    ")
// 	return string(o), err
// }

func cmdListBuild(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "List of knative build resources",
		Run: func(cmd *cobra.Command, args []string) {
			b.Namespace = client.Namespace
			if len(args) == 0 {
				output, err := b.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			} else {
				b.Name = args[0]
				output, err := b.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := bt.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			} else {
				bt.Name = args[0]
				output, err := bt.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := c.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			} else {
				c.Name = args[0]
				output, err := c.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := s.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			} else {
				s.Name = args[0]
				output, err := s.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := cf.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(output)
			} else {
				cf.Name = args[0]
				output, err := cf.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := r.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(output)
			} else {
				r.Name = args[0]
				output, err := r.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
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
				output, err := rt.List(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(output)
			} else {
				rt.Name = args[0]
				output, err := rt.Describe(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}
