package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"

	buildApi "github.com/knative/build/pkg/client/clientset/versioned"
	servingApi "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	debug     bool
	cfgFile   string
	namespace string
	output    string
	log       logrus.Logger
	build     *buildApi.Clientset
	serving   *servingApi.Clientset
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tm",
	Short: "Triggermesh CLI",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "k8s config file (default is ~/.kube/config)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "User namespace")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format")
}

func initConfig() {
	log = *logrus.New()
	log.Out = os.Stdout

	logFormat := new(logrus.TextFormatter)
	logFormat.TimestampFormat = "2006-01-02 15:04:05"
	logFormat.FullTimestamp = true
	log.Formatter = logFormat

	if debug {
		log.Level = logrus.DebugLevel
	}

	if len(cfgFile) == 0 {
		usr, err := user.Current()
		if err != nil {
			log.Panicln(err)
		}
		cfgFile = usr.HomeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		log.Panicln(err)
	}

	build, err = buildApi.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}
	serving, err = servingApi.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}
}

func toJSON(v interface{}) string {
	o, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		log.Errorln(err)
	}
	return string(o)
}

func toYAML(v interface{}) string {
	o, err := yaml.Marshal(v)
	if err != nil {
		log.Errorln(err)
	}
	return string(o)
}
