package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gosuri/uitable"
	buildApi "github.com/knative/build/pkg/client/clientset/versioned"
	servingApi "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	confPath = "/.tm/config.json"
)

var (
	debug     bool
	cfgFile   string
	namespace string
	output    string
	log       logrus.Logger
	core      *kubernetes.Clientset
	build     *buildApi.Clientset
	serving   *servingApi.Clientset
	table     *uitable.Table
)

type confStruct struct {
	Contexts []struct {
		Context struct {
			Cluster   string `json:"cluster"`
			Namespace string `json:"namespace"`
		} `json:"context"`
		Name string `json:"name"`
	} `json:"contexts"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "tm",
	Short:   "Triggermesh CLI",
	Version: "0.0.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("k8s config file (default is $HOME/%s)", confPath))
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "User namespace")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format")
}

func username() string {
	jsonFile, err := os.Open(cfgFile)
	if err != nil {
		log.Panicln(err)
	}
	defer jsonFile.Close()

	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panicln(err)
	}

	var conf confStruct
	if err := json.Unmarshal(body, &conf); err != nil {
		log.Panicln(err)
	}
	for _, v := range conf.Contexts {
		if v.Context.Cluster == "triggermesh" {
			return v.Context.Namespace
		}
	}
	return ""
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

	table = uitable.New()
	table.Wrap = true
	table.MaxColWidth = 50

	usr, err := user.Current()
	if err != nil {
		log.Panicln(err)
	}
	tmHome := filepath.Dir(usr.HomeDir + confPath)
	if _, err := os.Stat(tmHome); os.IsNotExist(err) {
		if err := os.MkdirAll(tmHome, 0755); err != nil {
			log.Panicln(err)
		}
	}

	if len(cfgFile) == 0 {
		cfgFile = usr.HomeDir + confPath
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		log.Panicln(err)
	}

	if len(namespace) == 0 {
		namespace = username()
	}

	build, err = buildApi.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}
	serving, err = servingApi.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}
	core, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}
}
