package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

func newDeployCmd(clientset *client.ConfigSet) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy knative resource",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := s.DeployYAML(YAML, args, clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployCmd.Flags().StringVarP(&YAML, "from", "f", "serverless.yaml", "Deploy functions defined in yaml")
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
			s.Name = args[0]
			s.Namespace = client.Namespace
			s.Registry = client.Registry
			output, err := s.Deploy(clientset)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(output)
		},
	}

	deployServiceCmd.Flags().StringVarP(&s.Source, "from", "f", "", "Service source to deploy: local folder with sources, git repository or docker image")
	deployServiceCmd.Flags().StringVar(&s.Revision, "revision", "master", "Git revision (branch, tag, commit SHA or ref)")
	deployServiceCmd.Flags().StringVar(&s.Buildtemplate, "build-template", "", "Existing buildtemplate name, local path or URL to buildtemplate yaml file")
	deployServiceCmd.Flags().StringVar(&s.RegistrySecret, "registry-secret", "", "Name of k8s secret to use in buildtemplate as registry auth json")
	deployServiceCmd.Flags().StringVar(&s.ResultImageTag, "tag", "latest", "Image tag to build")
	deployServiceCmd.Flags().StringVar(&s.PullPolicy, "image-pull-policy", "Always", "Image pull policy")
	deployServiceCmd.Flags().StringVar(&s.BuildTimeout, "build-timeout", "10m", "Service image build timeout")
	deployServiceCmd.Flags().IntVar(&s.Concurrency, "concurrency", 0, "Number of concurrent events per container: 0 - multiple events, 1 - single event, N - particular number of events")
	deployServiceCmd.Flags().StringSliceVar(&s.BuildArgs, "build-argument", []string{}, "Buildtemplate arguments")
	deployServiceCmd.Flags().StringSliceVar(&s.EnvSecrets, "env-secret", []string{}, "Name of k8s secrets to populate pod environment variables")
	deployServiceCmd.Flags().StringSliceVarP(&s.Labels, "label", "l", []string{}, "Service labels")
	deployServiceCmd.Flags().StringSliceVarP(&s.Env, "env", "e", []string{}, "Environment variables of the service, eg. `--env foo=bar`")
	return deployServiceCmd
}

func cmdDeployBuild(clientset *client.ConfigSet) *cobra.Command {
	deployBuildCmd := &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Args:    cobra.ExactArgs(1),
		Short:   "Deploy knative build",
		Example: "tm deploy build foo-builder --source git-repo --buildtemplate kaniko --args IMAGE=knative-local-registry:5000/foo-image",
		Run: func(cmd *cobra.Command, args []string) {
			b.Name = args[0]
			b.Namespace = client.Namespace
			if err := b.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployBuildCmd.Flags().StringVar(&b.Source, "source", "", "Git URL to get sources from")
	deployBuildCmd.Flags().StringVar(&b.Revision, "revision", "master", "Git source revision")
	deployBuildCmd.Flags().StringVar(&b.Buildtemplate, "buildtemplate", "", "Buildtemplate name to use with build")
	deployBuildCmd.Flags().StringVar(&b.Step, "step", "", "Build step (container) to run on provided source")
	deployBuildCmd.Flags().StringVar(&b.Image, "image", "", "Image for build step")
	deployBuildCmd.Flags().StringSliceVar(&b.Command, "command", []string{}, "Build step (container) command")
	deployBuildCmd.Flags().StringSliceVar(&b.Args, "args", []string{}, "Build arguments")
	deployBuildCmd.MarkFlagRequired("source")
	return deployBuildCmd
}

func cmdDeployBuildTemplate(clientset *client.ConfigSet) *cobra.Command {
	deployBuildTemplateCmd := &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates", "bldtmpl"},
		Short:   "Deploy knative build template",
		Example: "tm -n default deploy buildtemplate -f https://raw.githubusercontent.com/triggermesh/nodejs-runtime/master/knative-build-template.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			bt.Namespace = client.Namespace
			if _, err := bt.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}

	deployBuildTemplateCmd.Flags().StringVarP(&bt.File, "from", "f", "", "Local path or URL to buildtemplate yaml file")
	deployBuildTemplateCmd.Flags().StringVar(&bt.RegistrySecret, "credentials", "", "Name of k8s secret to use in buildtemplate as registry auth json")
	return deployBuildTemplateCmd
}

func cmdDeployChannel(clientset *client.ConfigSet) *cobra.Command {
	deployChannelCmd := &cobra.Command{
		Use:     "channel",
		Aliases: []string{"channels"},
		Args:    cobra.ExactArgs(1),
		Short:   "Deploy knative eventing channel",
		Run: func(cmd *cobra.Command, args []string) {
			c.Name = args[0]
			c.Namespace = client.Namespace
			if err := c.Deploy(clientset); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Channel deployment started")
		},
	}
	deployChannelCmd.Flags().StringVarP(&c.Provisioner, "provisioner", "p", "in-memory-channel", "Channel provisioner")
	return deployChannelCmd
}
