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

// listBuildsCmd represents the builds command
var listBuildsCmd = &cobra.Command{
	Use:   "builds",
	Short: "List of knative build resources",
	Run: func(cmd *cobra.Command, args []string) {
		l, err := build.BuildV1alpha1().Builds(namespace).List(metav1.ListOptions{})
		if err != nil {
			log.Error(err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		log.Debugf("+%v\n", l)

		fmt.Fprintln(w, "Name\tNamespace\t")
		for _, build := range l.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", build.Name, build.Namespace))
		}
		w.Flush()
	},
}

// servicesCmd represents the builds command
var listServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List of knative service resources",
	Run: func(cmd *cobra.Command, args []string) {
		list, err := serving.ServingV1alpha1().Services(namespace).List(metav1.ListOptions{})
		if err != nil {
			log.Errorln(err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

		log.Debugf("+%v\n", list)

		fmt.Fprintln(w, "Name\tNamespace\t")
		for _, service := range list.Items {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", service.Name, service.Namespace))
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(listBuildsCmd)
	getCmd.AddCommand(listServicesCmd)
}
