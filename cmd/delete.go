package cmd

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources from k8s cluster",
}

// deleteServiceCmd represents the builds command
var deleteServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Delete knative service resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := serving.ServingV1alpha1().Services(namespace).Delete(args[0], &metav1.DeleteOptions{}); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Service is being deleted")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteServiceCmd)
}
