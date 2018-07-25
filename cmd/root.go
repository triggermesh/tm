package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/knative/build/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfgFile   string
	namespace string
	k8sclient *versioned.Clientset
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tm",
	Short: "Triggermesh CLI",
	Long:  `Triggermesh CLI long description`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "k8s config file (default is ~/.kube/config)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "User namespace")
}

func initConfig() {
	if len(cfgFile) == 0 {
		usr, err := user.Current()
		if err != nil {
			panic(err.Error())
		}
		cfgFile = usr.HomeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		panic(err.Error())
	}

	k8sclient, err = versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}
