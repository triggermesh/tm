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
		listBuilds(args)
	},
}

// servicesCmd represents the builds command
var listServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List of knative service resources",
	Run: func(cmd *cobra.Command, args []string) {
		listServices(args)
	},
}

// listRoutesCmd represents the builds command
var listRoutesCmd = &cobra.Command{
	Use:   "routes",
	Short: "List of knative routes resources",
	Run: func(cmd *cobra.Command, args []string) {
		listRoutes(args)
	},
}

var listRevisionsCmd = &cobra.Command{
	Use:   "revisions",
	Short: "List of knative revision resources",
	Run: func(cmd *cobra.Command, args []string) {
		listRevisions(args)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(listBuildsCmd)
	getCmd.AddCommand(listServicesCmd)
	getCmd.AddCommand(listRoutesCmd)
	getCmd.AddCommand(listRevisionsCmd)
}

func listBuilds(args []string) {
	list, err := build.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return
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
}

func listServices(args []string) {
	list, err := serving.ServingV1alpha1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln(err)
		return
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
}

func listRoutes(args []string) {
	list, err := serving.ServingV1alpha1().Routes(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln(err)
		return
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
}

func listRevisions(args []string) {
	list, err := serving.ServingV1alpha1().Revisions(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln(err)
		return
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
}
