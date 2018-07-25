package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve resources from k8s cluster",
}

func init() {
	rootCmd.AddCommand(getCmd)
}
