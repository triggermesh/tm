package cmd

import (
	"github.com/spf13/cobra"
)

var (
	traffic int
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource parameters",
}

var setRouteCmd = &cobra.Command{
	Use:   "route",
	Short: "Configure service route",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setPercentage(args)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setRouteCmd)
	setRouteCmd.Flags().IntVarP(&traffic, "traffic", "t", 0, "Set traffic percentage")
	setRouteCmd.MarkFlagRequired("traffic")
}

func setPercentage(args []string) {
	// fmt.Printf("Route: %s, percentage: %d\n", args[0], traffic)
}
