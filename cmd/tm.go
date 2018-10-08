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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/cmd/deploy"
	"github.com/triggermesh/tm/cmd/describe"
	"github.com/triggermesh/tm/cmd/get"
	"github.com/triggermesh/tm/cmd/set"
	"github.com/triggermesh/tm/pkg/client"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	debug     bool
	cfgFile   string
	namespace string
	registry  string
	output    string
	clientset client.ClientSet
	log       *logrus.Logger
	err       error
)

// tmCmd represents the base command when called without any subcommands
var tmCmd = &cobra.Command{
	Use:     "tm",
	Short:   "Triggermesh CLI",
	Version: "0.0.3",
}

func Execute() {
	if err := tmCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	tmCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	tmCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "k8s config file")
	tmCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "User namespace")
	tmCmd.PersistentFlags().StringVar(&registry, "registry-host", "registry.munu.io", "User namespace")
	tmCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format")

	tmCmd.AddCommand(set.NewSetCmd(&clientset, log))
	tmCmd.AddCommand(deploy.NewDeployCmd(&clientset, log))
	tmCmd.AddCommand(delete.NewDeleteCmd(&clientset, log))
	tmCmd.AddCommand(get.NewGetCmd(&clientset, log, &output))
	tmCmd.AddCommand(describe.NewDescribeCmd(&clientset, log, &output))
}

func initConfig() {
	log = logrus.New()
	log.Out = os.Stdout

	logFormat := new(logrus.TextFormatter)
	logFormat.TimestampFormat = "2006-01-02 15:04:05"
	logFormat.FullTimestamp = true
	log.Formatter = logFormat

	if debug {
		log.Level = logrus.DebugLevel
	}

	clientset, err = client.NewClient(cfgFile, namespace, registry)
	if err != nil {
		log.Fatalln(err)
	}
}
