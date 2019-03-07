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

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/resources/build"
	"github.com/triggermesh/tm/pkg/resources/buildtemplate"
	"github.com/triggermesh/tm/pkg/resources/channel"
	"github.com/triggermesh/tm/pkg/resources/clusterbuildtemplate"
	"github.com/triggermesh/tm/pkg/resources/configuration"
	"github.com/triggermesh/tm/pkg/resources/credential"
	"github.com/triggermesh/tm/pkg/resources/revision"
	"github.com/triggermesh/tm/pkg/resources/route"
	"github.com/triggermesh/tm/pkg/resources/service"

	// Required for configs with gcp auth provider
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	version   string
	err       error
	kubeConf  string
	clientset client.ConfigSet

	YAML string
	b    build.Build
	c    channel.Channel
	s    service.Service
	r    revision.Revision
	rt   route.Route
	cf   configuration.Configuration
	gc   credential.GitCreds
	rc   credential.RegistryCreds
	bt   buildtemplate.Buildtemplate
	cbt  clusterbuildtemplate.ClusterBuildtemplate
)

// tmCmd represents the base command when called without any subcommands
var tmCmd = &cobra.Command{
	Use:     "tm",
	Short:   "Triggermesh CLI",
	Version: version,
}

func Execute() {
	if err := tmCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	tmCmd.PersistentFlags().StringVar(&kubeConf, "config", "", "k8s config file")
	tmCmd.PersistentFlags().StringVarP(&client.Namespace, "namespace", "n", "", "User namespace")
	tmCmd.PersistentFlags().StringVar(&client.Registry, "registry-host", "knative.registry.svc.cluster.local", "Docker registry host address")
	tmCmd.PersistentFlags().StringVarP(&client.Output, "output", "o", "", "Output format")
	tmCmd.PersistentFlags().BoolVar(&client.Wait, "wait", false, "Wait for the operation to complete")
	tmCmd.PersistentFlags().BoolVar(&client.Dry, "dry", false, "Do not create k8s objects, just print its structure")

	tmCmd.AddCommand(versionCmd)
	tmCmd.AddCommand(newDeployCmd(&clientset))
	tmCmd.AddCommand(newDeleteCmd(&clientset))
	tmCmd.AddCommand(newSetCmd(&clientset))
	tmCmd.AddCommand(newGetCmd(&clientset))
	// Describe is an alias for "get" command
	// tmCmd.AddCommand(newDescribeCmd(&clientset))
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of tm CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s, version %s\n", tmCmd.Short, version)
	},
}

func initConfig() {
	confPath := client.ConfigPath(kubeConf)
	if clientset, err = client.NewClient(confPath); err != nil {
		log.Fatalln(err)
	}
}
