package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var (
	YAML string
	// build         build.Build
	// channel       channel.Channel
	service       service.Service
	buildtemplate buildtemplate.Buildtemplate
)

func newDeployCmd(clientset *client.ConfigSet) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy knative resource",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := service.DeployYAML(YAML, args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployCmd.Flags().StringVarP(&YAML, "from", "f", "serverless.yaml", "Deploy functions defined in yaml")
	deployCmd.AddCommand(cmdDeployService(clientset))
	// deployCmd.AddCommand(cmdDeployChannel(clientset))
	// deployCmd.AddCommand(cmdDeployBuild(clientset))
	// deployCmd.AddCommand(cmdDeployBuildTemplate(clientset))

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
			service.Namespace = client.Namespace
			service.Registry = client.Registry
			output, err := service.Deploy(clientset)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(output)
		},
	}

	deployServiceCmd.Flags().StringVarP(&service.Source, "from", "f", "", "Service source to deploy: local folder with sources, git repository or docker image")
	deployServiceCmd.Flags().StringVar(&service.Revision, "revision", "master", "Git revision (branch, tag, commit SHA or ref)")
	deployServiceCmd.Flags().StringVar(&service.Buildtemplate, "build-template", "", "Existing buildtemplate name, local path or URL to buildtemplate yaml file")
	deployServiceCmd.Flags().StringVar(&service.RegistrySecret, "registry-secret", "", "Name of k8s secret to use in buildtemplate as registry auth json")
	deployServiceCmd.Flags().StringVar(&service.ResultImageTag, "tag", "latest", "Image tag to build")
	deployServiceCmd.Flags().StringVar(&service.PullPolicy, "image-pull-policy", "Always", "Image pull policy")
	deployServiceCmd.Flags().StringVar(&service.BuildTimeout, "build-timeout", "10m", "Service image build timeout")
	deployServiceCmd.Flags().IntVar(&service.Concurrency, "concurrency", 0, "Number of concurrent events per container: 0 - multiple events, 1 - single event, N - particular number of events")
	deployServiceCmd.Flags().StringSliceVar(&service.BuildArgs, "build-argument", []string{}, "Buildtemplate arguments")
	deployServiceCmd.Flags().StringSliceVar(&service.EnvSecrets, "env-secret", []string{}, "Name of k8s secrets to populate pod environment variables")
	deployServiceCmd.Flags().StringSliceVarP(&service.Labels, "label", "l", []string{}, "Service labels")
	deployServiceCmd.Flags().StringSliceVarP(&service.Env, "env", "e", []string{}, "Environment variables of the service, eg. `--env foo=bar`")

	return deployServiceCmd
}
