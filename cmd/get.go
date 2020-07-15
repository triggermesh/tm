/*
Copyright (c) 2020 TriggerMesh Inc.

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
				if len(list.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No channels found\n")
					return
				}
				clientset.Printer.PrintTable(c.GetTable(list))
				return
			}
			c.Name = args[0]
			channel, err := c.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(c.GetObject(channel))
		},
	}
}

func cmdListService(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "service",
		Aliases: []string{"services"},
		Short:   "List of knative service resources",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s.Namespace = client.Namespace
			if len(args) == 0 {
				list, err := s.List(clientset)
				if err != nil {
					clientset.Log.Fatalln(err)
				}
				if len(list.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No services found\n")
					return
				}
				clientset.Printer.PrintTable(s.GetTable(list))
				return
			}
			s.Name = args[0]
			service, err := s.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(s.GetObject(service))
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
				if len(list.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No configurations found\n")
					return
				}
				clientset.Printer.PrintTable(cf.GetTable(list))
				return
			}
			cf.Name = args[0]
			configuration, err := cf.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(cf.GetObject(configuration))
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
				if len(list.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No revisions found\n")
					return
				}
				clientset.Printer.PrintTable(r.GetTable(list))
				return
			}
			r.Name = args[0]
			revision, err := r.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(r.GetObject(revision))
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
				clientset.Printer.PrintTable(rt.GetTable(list))
				return
			}
			rt.Name = args[0]
			route, err := rt.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(rt.GetObject(route))
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
				clientset.Printer.PrintTable(t.GetTable(list))
				return
			}
			t.Name = args[0]
			task, err := t.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(t.GetObject(task))
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
				clientset.Printer.PrintTable(tr.GetTable(list))
				return
			}
			tr.Name = args[0]
			taskrun, err := tr.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(tr.GetObject(taskrun))
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
				clientset.Printer.PrintTable(plr.GetTable(list))
				return
			}
			plr.Name = args[0]
			pipelineResource, err := plr.Get(clientset)
			if err != nil {
				clientset.Log.Fatalln(err)
			}
			clientset.Printer.PrintObject(plr.GetObject(pipelineResource))
		},
	}
}
