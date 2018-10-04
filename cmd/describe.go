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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Details of knative resources",
}

var describeServiceCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"services"},
	Short:   "Knative service details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeService(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listServices()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describeConfigurationCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"configurations"},
	Short:   "Knative service configuration details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeConfiguration(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listConfigurations()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describeRevisionCmd = &cobra.Command{
	Use:     "revision",
	Aliases: []string{"revisions"},
	Short:   "Knative revision details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeRevision(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listRevisions()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describeRouteCmd = &cobra.Command{
	Use:     "route",
	Aliases: []string{"routes"},
	Short:   "Knative service route details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeRoute(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listRoutes()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describePodCmd = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"pods"},
	Short:   "Pod details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describePod(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listPods()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describeBuildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"builds"},
	Short:   "Build details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeBuild(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listBuilds()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

var describeBuildTemplateCmd = &cobra.Command{
	Use:     "buildtemplate",
	Aliases: []string{"buildtemplates"},
	Short:   "Buildtemplate details",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			output, err := describeBuildTemplate(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		} else {
			output, err := listBuildTemplates()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.AddCommand(describeServiceCmd)
	describeCmd.AddCommand(describeConfigurationCmd)
	describeCmd.AddCommand(describeRevisionCmd)
	describeCmd.AddCommand(describeRouteCmd)
	describeCmd.AddCommand(describePodCmd)
	describeCmd.AddCommand(describeBuildCmd)
	describeCmd.AddCommand(describeBuildTemplateCmd)

}

func describeService(args []string) ([]byte, error) {
	service, err := serving.ServingV1alpha1().Services(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(service)
	}
	return json.MarshalIndent(service, "", "	")
}

func describeConfiguration(args []string) ([]byte, error) {
	configuration, err := serving.ServingV1alpha1().Configurations(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(configuration)
	}
	return json.MarshalIndent(configuration, "", "	")
}

func describeRevision(args []string) ([]byte, error) {
	revisions, err := serving.ServingV1alpha1().Revisions(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(revisions)
	}
	return json.MarshalIndent(revisions, "", "	")
}

func describeRoute(args []string) ([]byte, error) {
	routes, err := serving.ServingV1alpha1().Routes(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(routes)
	}
	return json.MarshalIndent(routes, "", "	")
}

func describePod(args []string) ([]byte, error) {
	pods, err := core.CoreV1().Pods(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(pods)
	}
	return json.MarshalIndent(pods, "", "	")
}

func describeBuild(args []string) ([]byte, error) {
	build, err := build.BuildV1alpha1().Builds(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(build)
	}
	return json.MarshalIndent(build, "", "	")
}

func describeBuildTemplate(args []string) ([]byte, error) {
	buildtemplate, err := build.BuildV1alpha1().BuildTemplates(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(buildtemplate)
	}
	return json.MarshalIndent(buildtemplate, "", "	")
}
