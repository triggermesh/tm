package set

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

var r Route
var c Credentials

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource parameters",
}

func NewSetCmd(clientset *client.ClientSet) *cobra.Command {
	setCmd.AddCommand(cmdSetRoutes(clientset))
	setCmd.AddCommand(cmdSetCredentials(clientset))
	return setCmd
}

func cmdSetRoutes(clientset *client.ClientSet) *cobra.Command {
	setRoutesCmd := &cobra.Command{
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

	setRoutesCmd.Flags().StringSliceVarP(&r.Revisions, "revisions", "r", []string{}, "Set traffic percentage for revision")
	setRoutesCmd.Flags().StringSliceVarP(&r.Configs, "configurations", "c", []string{}, "Set traffic percentage for configuration")
	return setRoutesCmd
}

func cmdSetCredentials(clientset *client.ClientSet) *cobra.Command {
	setCredentialsCmd := &cobra.Command{
		Use:   "credentials",
		Short: "Set secret with registry credentials",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.SetCredentials(args, clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Routes successfully updated")
		},
	}

	setCredentialsCmd.Flags().StringVar(&c.Host, "registry", "", "Registry host address")
	setCredentialsCmd.Flags().StringVar(&c.Username, "username", "", "Registry username")
	setCredentialsCmd.Flags().StringVar(&c.Password, "password", "", "Registry password")
	return setCredentialsCmd
}
