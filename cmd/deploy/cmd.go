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
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var s Service
var b Buildtemplate

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy knative resource",
}

func NewDeployCmd(clientset *client.ClientSet) *cobra.Command {
	deployCmd.AddCommand(cmdDeployService(clientset))
	deployCmd.AddCommand(cmdDeployBuildTemplate(clientset))
	return deployCmd
}

func cmdDeployService(clientset *client.ClientSet) *cobra.Command {
	deployServiceCmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Deploy knative service",
		Args:    cobra.ExactArgs(1),
		Example: "tm -n default deploy service foo --build-template kaniko --build-argument SKIP_TLS_VERIFY=true --from-image gcr.io/google-samples/hello-app:1.0",
		Run: func(cmd *cobra.Command, args []string) {
			if err := s.DeployService(args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployServiceCmd.Flags().StringVar(&s.From.Image.URL, "from-image", "", "Image to deploy")
	deployServiceCmd.Flags().StringVar(&s.From.Repository.URL, "from-source", "", "Git source URL to deploy")
	deployServiceCmd.Flags().StringVar(&s.From.Repository.Revision, "revision", "master", "May be used with \"--from-source\" flag: git revision (branch, tag, commit SHA or ref) to clone")
	deployServiceCmd.Flags().StringVar(&s.From.Path, "from-file", "", "Local file path to deploy")
	deployServiceCmd.Flags().StringVar(&s.From.URL, "from-url", "", "File source URL to deploy")
	deployServiceCmd.Flags().StringVar(&s.Buildtemplate, "build-template", "kaniko", "Build template to use with service")
	deployServiceCmd.Flags().StringVar(&s.ResultImageTag, "tag", "latest", "Image tag to build")
	deployServiceCmd.Flags().StringVar(&s.PullPolicy, "image-pull-policy", "Always", "Image pull policy")
	deployServiceCmd.Flags().StringSliceVar(&s.BuildArgs, "build-argument", []string{}, "Image tag to build")
	deployServiceCmd.Flags().StringSliceVarP(&s.Labels, "label", "l", []string{}, "Service labels")
	deployServiceCmd.Flags().StringSliceVarP(&s.Env, "env", "e", []string{}, "Environment variables of the service, eg. `--env foo=bar`")

	return deployServiceCmd
}

func cmdDeployBuildTemplate(clientset *client.ClientSet) *cobra.Command {
	deployBuildTemplateCmd := &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtempalte", "bldtmpl"},
		Short:   "Deploy knative build template",
		Example: "tm -n default deploy buildtemplate --from-url https://raw.githubusercontent.com/triggermesh/nodejs-runtime/master/knative-build-template.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			if err := b.DeployBuildTemplate(args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployBuildTemplateCmd.Flags().StringVar(&b.URL, "from-url", "", "Build template yaml URL")
	deployBuildTemplateCmd.Flags().StringVar(&b.Path, "from-file", "", "Local file path to deploy")

	return deployBuildTemplateCmd
}
