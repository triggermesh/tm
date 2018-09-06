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
		if err := deleteService(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Service is being deleted")
	},
}

var deleteConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Delete knative configuration resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteConfiguration(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Configuration is being deleted")
	},
}

var deleteBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Delete knative build resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteBuild(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Build is being deleted")
	},
}

var deleteBuildTemplateCmd = &cobra.Command{
	Use:   "buildtemplate",
	Short: "Delete knative buildtemplate resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteBuildTemplate(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Buildtemplate is being deleted")
	},
}

var deleteRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "Delete knative revision resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteRevision(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Revision is being deleted")
	},
}

var deleteRouteCmd = &cobra.Command{
	Use:   "route",
	Short: "Delete knative route resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteRoute(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Route is being deleted")
	},
}

var deletePodCmd = &cobra.Command{
	Use:   "pod",
	Short: "Delete kubernetes pod resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deletePod(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Info("Pod is being deleted")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteServiceCmd)
	deleteCmd.AddCommand(deleteConfigurationCmd)
	deleteCmd.AddCommand(deleteBuildCmd)
	deleteCmd.AddCommand(deleteBuildTemplateCmd)
	deleteCmd.AddCommand(deleteRevisionCmd)
	deleteCmd.AddCommand(deleteRouteCmd)
	deleteCmd.AddCommand(deletePodCmd)
}

func deletePod(args []string) error {
	return core.CoreV1().Pods(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteRoute(args []string) error {
	return serving.ServingV1alpha1().Routes(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteRevision(args []string) error {
	return serving.ServingV1alpha1().Revisions(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteBuildTemplate(args []string) error {
	return build.BuildV1alpha1().BuildTemplates(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteBuild(args []string) error {
	return build.BuildV1alpha1().Builds(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteConfiguration(args []string) error {
	return serving.ServingV1alpha1().Configurations(namespace).Delete(args[0], &metav1.DeleteOptions{})
}

func deleteService(args []string) error {
	return serving.ServingV1alpha1().Services(namespace).Delete(args[0], &metav1.DeleteOptions{})
}
