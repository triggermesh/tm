// Placeholder
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Not yet implemented",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pods list")
	},
}

func init() {
	getCmd.AddCommand(podsCmd)
}
