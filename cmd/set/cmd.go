package set

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var r Route

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource parameters",
}

func cmdSetRoutes(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:   "route",
		Short: "Configure service route",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := r.SetPercentage(args, clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Routes successfully updated")
		},
	}
}

func NewSetCmd(clientset *client.ClientSet) *cobra.Command {
	setCmd.AddCommand(cmdSetRoutes(clientset))
	setCmd.Flags().StringSliceVarP(&r.Revisions, "revisions", "r", []string{}, "Set traffic percentage for revision")
	setCmd.Flags().StringSliceVarP(&r.Configs, "configs", "c", []string{}, "Set traffic percentage for configuration")
	return setCmd
}
