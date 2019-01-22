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

package deploy

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var (
	service       Service
	build         Build
	buildtemplate Buildtemplate
	channel       Channel
	YAML          string
)

// NewDeployCmd returns deploy cobra command and its subcommands
func NewDeployCmd(clientset *client.ConfigSet) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy knative resource",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := service.DeployYAML(YAML, args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployCmd.Flags().StringVarP(&YAML, "file", "f", "serverless.yaml", "Deploy functions defined in yaml")
	deployCmd.Flags().BoolVarP(&service.Wait, "wait", "w", false, "Wait for each function deployment")

	deployCmd.AddCommand(cmdDeployService(clientset))
	deployCmd.AddCommand(cmdDeployChannel(clientset))
	deployCmd.AddCommand(cmdDeployBuild(clientset))
	deployCmd.AddCommand(cmdDeployBuildTemplate(clientset))

	return deployCmd
}

func cmdDeployService(clientset *client.ConfigSet) *cobra.Command {
	deployServiceCmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Deploy knative service",
		Args:    cobra.ExactArgs(1),
		Example: "tm -n default deploy service foo --build-template kaniko --from-image gcr.io/google-samples/hello-app:1.0",
		Run: func(cmd *cobra.Command, args []string) {
			service.Name = args[0]
			if err := service.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	// kept for back compatibility
	deployServiceCmd.Flags().StringVar(&service.Source, "from-path", "", "Deprecated, use `-f` flag instead")
	deployServiceCmd.Flags().StringVar(&service.Source, "from-image", "", "Deprecated, use `-f` flag instead")
	deployServiceCmd.Flags().StringVar(&service.Source, "from-source", "", "Deprecated, use `-f` flag instead")

	deployServiceCmd.Flags().StringVarP(&service.Source, "from", "f", "", "Service source to deploy: local folder with sources, git repository or docker image")
	deployServiceCmd.Flags().StringVar(&service.Revision, "revision", "master", "Git revision (branch, tag, commit SHA or ref)")
	deployServiceCmd.Flags().BoolVar(&service.Wait, "wait", false, "Wait for successful service deployment")
	deployServiceCmd.Flags().StringVar(&service.Buildtemplate, "build-template", "", "Existing buildtemplate name, local path or URL to buildtemplate yaml file")
	deployServiceCmd.Flags().StringVar(&service.RegistrySecret, "registry-secret", "", "Name of k8s secret to use in buildtemplate as registry auth json")
	deployServiceCmd.Flags().StringVar(&service.ResultImageTag, "tag", "latest", "Image tag to build")
	deployServiceCmd.Flags().StringVar(&service.PullPolicy, "image-pull-policy", "Always", "Image pull policy")
	deployServiceCmd.Flags().StringSliceVar(&service.BuildArgs, "build-argument", []string{}, "Buildtemplate arguments")
	deployServiceCmd.Flags().StringSliceVar(&service.EnvSecrets, "env-secret", []string{}, "Name of k8s secrets to populate pod environment variables")
	deployServiceCmd.Flags().StringSliceVarP(&service.Labels, "label", "l", []string{}, "Service labels")
	deployServiceCmd.Flags().StringSliceVarP(&service.Env, "env", "e", []string{}, "Environment variables of the service, eg. `--env foo=bar`")

	return deployServiceCmd
}

func cmdDeployBuildTemplate(clientset *client.ConfigSet) *cobra.Command {
	deployBuildTemplateCmd := &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates", "bldtmpl"},
		Short:   "Deploy knative build template",
		Example: "tm -n default deploy buildtemplate -f https://raw.githubusercontent.com/triggermesh/nodejs-runtime/master/knative-build-template.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := buildtemplate.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	// kept for back compatibility
	deployBuildTemplateCmd.Flags().StringVar(&buildtemplate.File, "from-url", "", "Deprecated, use `-f` flag instead")
	deployBuildTemplateCmd.Flags().StringVar(&buildtemplate.File, "from-file", "", "Deprecated, use `-f` flag instead")

	deployBuildTemplateCmd.Flags().StringVarP(&buildtemplate.File, "from", "f", "", "Local path or URL to buildtemplate yaml file")
	deployBuildTemplateCmd.Flags().StringVar(&buildtemplate.RegistrySecret, "credentials", "", "Name of k8s secret to use in buildtemplate as registry auth json")

	return deployBuildTemplateCmd
}

func cmdDeployBuild(clientset *client.ConfigSet) *cobra.Command {
	deployBuildCmd := &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Args:    cobra.ExactArgs(1),
		Short:   "Deploy knative build",
		Example: "tm deploy build foo-builder --source git-repo --buildtemplate kaniko --args IMAGE=knative-local-registry:5000/foo-image",
		Run: func(cmd *cobra.Command, args []string) {
			build.Name = args[0]
			if err := build.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployBuildCmd.Flags().StringVar(&build.Source, "source", "", "Git URL to get sources from")
	deployBuildCmd.Flags().StringVar(&build.Revision, "revision", "master", "Git source revision")
	deployBuildCmd.Flags().StringVar(&build.Buildtemplate, "buildtemplate", "", "Buildtemplate name to use with build")
	deployBuildCmd.Flags().StringVar(&build.Step, "step", "", "Build step (container) to run on provided source")
	deployBuildCmd.Flags().StringVar(&build.Image, "image", "", "Image for build step")
	deployBuildCmd.Flags().StringSliceVar(&build.Command, "command", []string{}, "Build step (container) command")
	deployBuildCmd.Flags().StringSliceVar(&build.Args, "args", []string{}, "Build arguments")
	deployBuildCmd.MarkFlagRequired("source")

	return deployBuildCmd
}

func cmdDeployChannel(clientset *client.ConfigSet) *cobra.Command {
	deployChannelCmd := &cobra.Command{
		Use:     "channel",
		Aliases: []string{"channels"},
		Args:    cobra.ExactArgs(1),
		Short:   "Deploy knative eventing channel",
		Run: func(cmd *cobra.Command, args []string) {
			channel.Name = args[0]
			if err := channel.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Channel deployment started")
		},
	}
	deployChannelCmd.Flags().StringVarP(&channel.Provisioner, "provisioner", "p", "in-memory-channel", "Channel provisioner")
	return deployChannelCmd
}
