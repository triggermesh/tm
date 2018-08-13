package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve resources from k8s cluster",
}

// TODO: single interface for resulting object

// listBuildsCmd represents the builds command
var listBuildsCmd = &cobra.Command{
	Use:   "builds",
	Short: "List of knative build resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listBuilds(args); err != nil {
			log.Errorln(err)
		}
	},
}

// servicesCmd represents the builds command
var listServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List of knative service resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listServices(args); err != nil {
			log.Errorln(err)
		}
	},
}

// listRoutesCmd represents the builds command
var listRoutesCmd = &cobra.Command{
	Use:   "routes",
	Short: "List of knative routes resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listRoutes(args); err != nil {
			log.Errorln(err)
		}
	},
}

var listRevisionsCmd = &cobra.Command{
	Use:   "revisions",
	Short: "List of knative revision resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listRevisions(args); err != nil {
			log.Errorln(err)
		}
	},
}

var listPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "List of pods",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listPods(args); err != nil {
			log.Errorln(err)
		}
	},
}

var listBuildTemplatesCmd = &cobra.Command{
	Use:   "buildtemplates",
	Short: "List of buildtemplates",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listBuildTemplates(args); err != nil {
			log.Errorln(err)
		}
	},
}

var listConfigurationsCmd = &cobra.Command{
	Use:   "configurations",
	Short: "List of configurations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listConfigurations(args); err != nil {
			log.Errorln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(listBuildsCmd)
	getCmd.AddCommand(listServicesCmd)
	getCmd.AddCommand(listRoutesCmd)
	getCmd.AddCommand(listRevisionsCmd)
	getCmd.AddCommand(listPodsCmd)
	getCmd.AddCommand(listBuildTemplatesCmd)
	getCmd.AddCommand(listConfigurationsCmd)
}

func listBuilds(args []string) error {
	list, err := build.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Build\tNamespace\t")
		for _, build := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", build.Name, build.Namespace))
		}
		w.Flush()
	}
	return err
}

func listServices(args []string) error {
	list, err := serving.ServingV1alpha1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Service\tNamespace\t")
		for _, service := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", service.Name, service.Namespace))
		}
		w.Flush()
	}
	return err
}

func listRoutes(args []string) error {
	list, err := serving.ServingV1alpha1().Routes(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Route\tNamespace\tTarget(%)")
		for _, route := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%v\t", route.Name, route.Namespace, route.Spec.Traffic))
		}
		w.Flush()
	}
	return err
}

func listRevisions(args []string) error {
	list, err := serving.ServingV1alpha1().Revisions(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Revision\tNamespace")
		for _, revision := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", revision.Name, revision.Namespace))
		}
		w.Flush()
	}
	return err
}

func listPods(args []string) error {
	list, err := core.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Pods\tNamespace")
		for _, pod := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", pod.Name, pod.Namespace))
		}
		w.Flush()
	}
	return err
}

func listBuildTemplates(args []string) error {
	list, err := build.BuildV1alpha1().BuildTemplates(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Template\tNamespace")
		for _, template := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", template.Name, template.Namespace))
		}
		w.Flush()
	}
	return err
}

func listConfigurations(args []string) error {
	list, err := serving.ServingV1alpha1().Configurations(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	switch output {
	case "json":
		fmt.Println(toJSON(list))
	case "yaml":
		fmt.Println(toYAML(list))
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		fmt.Fprintln(w, "Configuration\tNamespace")
		for _, configuration := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", configuration.Name, configuration.Namespace))
		}
		w.Flush()
	}
	return err
}
