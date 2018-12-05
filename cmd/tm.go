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
	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/cmd/deploy"
	"github.com/triggermesh/tm/cmd/describe"
	"github.com/triggermesh/tm/cmd/get"
	"github.com/triggermesh/tm/cmd/set"
	"github.com/triggermesh/tm/pkg/client"

	// Required for configs with gcp auth provider
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	debug     bool
	err       error
	cfgFile   string
	namespace string
	registry  string
	output    string
	clientset client.ConfigSet
)

// tmCmd represents the base command when called without any subcommands
var tmCmd = &cobra.Command{
	Use:     "tm",
	Short:   "Triggermesh CLI",
	Version: "0.0.6",
}

// Execute runs cobra CLI commands
func Execute() {
	if err := tmCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	tmCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "k8s config file")
	tmCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "User namespace")
	tmCmd.PersistentFlags().StringVar(&registry, "registry-host", "knative.registry.svc.cluster.local", "Docker registry host address")
	tmCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format")

	tmCmd.AddCommand(versionCmd)
	tmCmd.AddCommand(set.NewSetCmd(&clientset))
	tmCmd.AddCommand(deploy.NewDeployCmd(&clientset))
	tmCmd.AddCommand(delete.NewDeleteCmd(&clientset))
	tmCmd.AddCommand(get.NewGetCmd(&clientset))
	tmCmd.AddCommand(describe.NewDescribeCmd(&clientset))
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of tm CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s, version %s\n", tmCmd.Short, tmCmd.Version)
	},
}

func initConfig() {
	describe.Format(&output)
	get.Format(&output)

	if clientset, err = client.NewClient(cfgFile, namespace, registry); err != nil {
		log.Fatalln(err)
	}
}
