package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	knativeApi "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/lib/knative"
)

// buildsCmd represents the builds command
var buildsCmd = &cobra.Command{
	Use:   "builds",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

func init() {
	getCmd.AddCommand(buildsCmd)
}

func list() (*knativeApi.BuildList, error) {
	l, err := knative.GetBuildListByNamespace(k8sclient, namespace)
	if err != nil {
		return nil, err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)

	fmt.Fprintln(w, "Name\tNamespace\t")
	for _, build := range l.Items {
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", build.Name, build.Namespace))
	}
	w.Flush()
	return nil, nil
}
