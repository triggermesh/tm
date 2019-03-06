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

// NewDeleteCmd returns cobra Command with set of resource deletion subcommands
func newDeleteCmd(clientset *client.ConfigSet) *cobra.Command {
	var file string
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete knative resource",
		Run: func(cmd *cobra.Command, args []string) {
			s.Namespace = client.Namespace
			if err := s.DeleteYAML(file, args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deleteCmd.Flags().StringVarP(&file, "file", "f", "serverless.yaml", "Delete functions defined in yaml")

	deleteCmd.AddCommand(cmdDeleteConfiguration(clientset))
	deleteCmd.AddCommand(cmdDeleteBuildTemplate(clientset))
	deleteCmd.AddCommand(cmdDeleteRevision(clientset))
	deleteCmd.AddCommand(cmdDeleteService(clientset))
	deleteCmd.AddCommand(cmdDeleteBuild(clientset))
	deleteCmd.AddCommand(cmdDeleteRoute(clientset))
	deleteCmd.AddCommand(cmdDeleteChannel(clientset))
	return deleteCmd
}

func cmdDeleteBuild(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "Delete knative build resource",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b.Name = args[0]
			b.Namespace = client.Namespace
			if err := b.DeleteBuild(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Build is being deleted")
		},
	}
}

func cmdDeleteBuildTemplate(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates"},
		Short:   "Delete knative buildtemplate resource",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bt.Name = args[0]
			bt.Namespace = client.Namespace
			if err := bt.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("BuildTemplate is being deleted")
		},
	}
}

func cmdDeleteChannel(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "channel",
		Aliases: []string{"channels"},
		Short:   "Delete knative channel resource",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c.Name = args[0]
			c.Namespace = client.Namespace
			if err := c.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Channel is being deleted")
		},
	}
}

func cmdDeleteService(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "service",
		Short:   "Delete knative service resource",
		Aliases: []string{"services"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s.Name = args[0]
			s.Namespace = client.Namespace
			if err := s.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Service is being deleted")
		},
	}
}

func cmdDeleteConfiguration(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "configuration",
		Short:   "Delete knative configuration resource",
		Aliases: []string{"configurations"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cf.Name = args[0]
			cf.Namespace = client.Namespace
			if err := cf.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Configuration is being deleted")
		},
	}
}

func cmdDeleteRevision(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "revision",
		Short:   "Delete knative revision resource",
		Aliases: []string{"revisions"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			r.Name = args[0]
			r.Namespace = client.Namespace
			if err := r.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Revision is being deleted")
		},
	}
}

func cmdDeleteRoute(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "route",
		Short:   "Delete knative route resource",
		Aliases: []string{"routes"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rt.Name = args[0]
			rt.Namespace = client.Namespace
			if err := rt.Delete(clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Route is being deleted")
		},
	}
}
