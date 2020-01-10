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
	getCmd.AddCommand(cmdListClusterBuildTemplates(clientset))
	getCmd.AddCommand(cmdListConfigurations(clientset))
	getCmd.AddCommand(cmdListRevision(clientset))
	getCmd.AddCommand(cmdListRoute(clientset))
	getCmd.AddCommand(cmdListService(clientset))
	getCmd.AddCommand(cmdListChannels(clientset))
	getCmd.AddCommand(cmdListTasks(clientset))
	getCmd.AddCommand(cmdListTaskRuns(clientset))
	getCmd.AddCommand(cmdListPipelineResources(clientset))

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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				b.Name = args[0]
				build, err := b.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(build, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				bt.Name = args[0]
				buildtemplate, err := bt.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(buildtemplate, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListClusterBuildTemplates(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "clusterbuildtemplate",
		Aliases: []string{"cbuildtemplates", "cbuildtemplate", "clusterbuildtemplates"},
		Short:   "List of clusterbuildtemplates",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				list, err := cbt.List(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				cbt.Name = args[0]
				cbuildtemplate, err := cbt.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(cbuildtemplate, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				c.Name = args[0]
				channel, err := c.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(channel, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				s.Name = args[0]
				service, err := s.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(service, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				cf.Name = args[0]
				configuration, err := cf.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(configuration, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				r.Name = args[0]
				revision, err := r.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(revision, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
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
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				rt.Name = args[0]
				route, err := rt.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(route, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListTasks(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "List of tekton task resources",
		Run: func(cmd *cobra.Command, args []string) {
			t.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := t.List(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				t.Name = args[0]
				task, err := t.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(task, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListTaskRuns(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "taskrun",
		Aliases: []string{"taskruns"},
		Short:   "List of tekton TaskRun resources",
		Run: func(cmd *cobra.Command, args []string) {
			tr.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := tr.List(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				tr.Name = args[0]
				taskrun, err := tr.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(taskrun, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}

func cmdListPipelineResources(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "pipelineresource",
		Aliases: []string{"pipelineresources"},
		Short:   "List of tekton PipelineResources resources",
		Run: func(cmd *cobra.Command, args []string) {
			plr.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := plr.List(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.List(list.Items)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			} else {
				plr.Name = args[0]
				pipelineResource, err := plr.Get(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				data, err = output.Describe(pipelineResource, client.Output)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
			}
			fmt.Println(data)
		},
	}
}
