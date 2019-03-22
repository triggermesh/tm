package file

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// Definition represents serverless.yaml file structure
type Definition struct {
	Service     string
	Description string
	Provider    struct {
		Name           string
		Registry       string
		RegistrySecret string `yaml:"registry-secret"`
		PullPolicy     string `yaml:"pull-policy"`
		Namespace      string
		Runtime        string
		Buildtimeout   string
		Environment    map[string]string
		EnvSecrets     []string `yaml:"env-secrets"`
		Annotations    map[string]string
	}
	Repository string
	Functions  map[string]Function
	Include    []string
}

// Function describes function definition in serverless format
type Function struct {
	Handler     string
	Source      string
	Runtime     string
	Concurrency int
	Buildargs   []string
	Description string
	Labels      []string
	Environment map[string]string
	EnvSecrets  []string `yaml:"env-secrets"`
	Annotations map[string]string
	Events      []map[string]interface{}
}

type Schedule struct {
	Rate string
	Data string
}

// ParseServerless accepts serverless yaml file path and returns decoded structure
func ParseManifest(path string) (Definition, error) {
	var f Definition
	if _, err := os.Stat(path); err != nil {
		return f, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return f, err
	}
	f.Repository = filepath.Base(filepath.Dir(path))
	err = yaml.Unmarshal(data, &f)

	return f, err
}
