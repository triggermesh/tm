package file

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
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
	Revision    string
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

var Aos = afero.NewOsFs()

// ParseServerless accepts serverless yaml file path and returns decoded structure
func ParseManifest(path string) (Definition, error) {
	var definition Definition

	exists, err := afero.Exists(Aos, path)

	if !exists || err != nil {
		return definition, errors.New("could not find manifest file")
	}

	data, err := afero.ReadFile(Aos, path)

	if err != nil {
		return definition, err
	}

	definition.Repository = filepath.Base(filepath.Dir(path))
	err = yaml.UnmarshalStrict(data, &definition)

	return definition, err
}

func (definition Definition) Validate() error {
	if definition.Provider.Name != "" && definition.Provider.Name != "triggermesh" {
		return fmt.Errorf("%s provider is not supported", definition.Provider.Name)
	}

	if len(definition.Service) == 0 {
		return errors.New("Service name can't be empty")
	}

	return nil
}
