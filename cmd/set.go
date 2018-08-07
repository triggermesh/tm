package cmd

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func setPercentage(args []string) {
	routes, err := serving.ServingV1alpha1().Routes(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		log.Errorln(err)
		return
	}
	for i, route := range routes.Spec.Traffic {
		route.Percent = traffic
		routes.Spec.Traffic[i] = route
	}
	routes, err = serving.ServingV1alpha1().Routes(namespace).Update(routes)
	if err != nil {
		log.Errorln(err)
		return
	}
	log.Debugln(routes.Spec.Traffic)
	log.Infoln("Routes successfully updated")
}
