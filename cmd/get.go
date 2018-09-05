package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve resources from k8s cluster",
}

// TODO: single interface for resulting object

// servicesCmd represents the builds command
var listServicesCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"services"},
	Short:   "List of knative service resources",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listServices()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeService(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

// listBuildsCmd represents the builds command
var listBuildsCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"builds"},
	Short:   "List of knative build resources",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listBuilds()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeBuild(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

// listRoutesCmd represents the builds command
var listRoutesCmd = &cobra.Command{
	Use:     "route",
	Aliases: []string{"routes"},
	Short:   "List of knative routes resources",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listRoutes()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeRoute(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

var listRevisionsCmd = &cobra.Command{
	Use:     "revision",
	Aliases: []string{"revisions"},
	Short:   "List of knative revision resources",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listRevisions()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeRevision(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

var listPodsCmd = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"pods"},
	Short:   "List of pods",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listPods()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describePod(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

var listBuildTemplatesCmd = &cobra.Command{
	Use:     "buildtemplate",
	Aliases: []string{"buildtemplates"},
	Short:   "List of buildtemplates",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listBuildTemplates()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeBuildTemplate(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

var listConfigurationsCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"configurations"},
	Short:   "List of configurations",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output, err := listConfigurations()
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(output)

		} else {
			output, err := describeConfiguration(args)
			if err != nil {
				log.Errorln(err)
			}
			fmt.Println(string(output))
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(listServicesCmd)
	getCmd.AddCommand(listBuildsCmd)
	getCmd.AddCommand(listRoutesCmd)
	getCmd.AddCommand(listRevisionsCmd)
	getCmd.AddCommand(listPodsCmd)
	getCmd.AddCommand(listBuildTemplatesCmd)
	getCmd.AddCommand(listConfigurationsCmd)
}

func format(v interface{}) (string, error) {
	switch output {
	case "json":
		o, err := json.MarshalIndent(v, "", "    ")
		return string(o), err
	case "yaml":
		o, err := yaml.Marshal(v)
		return string(o), err
	}
	return "", nil
}

func listBuilds() (string, error) {
	list, err := build.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "BUILD")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}

func listServices() (string, error) {
	list, err := serving.ServingV1alpha1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "SERVICE")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}

func listRoutes() (string, error) {
	list, err := serving.ServingV1alpha1().Routes(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "ROUTE", "TARGETS")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name, item.Spec.Traffic)
		}
		return table.String(), err
	}
	return format(list)
}

func listRevisions() (string, error) {
	list, err := serving.ServingV1alpha1().Revisions(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "REVISION")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}

func listPods() (string, error) {
	list, err := core.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "POD")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}

func listBuildTemplates() (string, error) {
	list, err := build.BuildV1alpha1().BuildTemplates(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "BUILDTEMPLATE")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}

func listConfigurations() (string, error) {
	list, err := serving.ServingV1alpha1().Configurations(namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "CONFIGURATION")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return format(list)
}
