package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildsCmd represents the builds command
var buildsCmd = &cobra.Command{
	Use:   "builds",
	Short: "List of knative build resources",
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

func init() {
	getCmd.AddCommand(buildsCmd)
}

func list() {
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
}
