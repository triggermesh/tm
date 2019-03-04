package service

import "encoding/json"

// Service represents knative service structure
type Service struct {
	Name           string
	Namespace      string
	Registry       string
	Source         string
	Revision       string
	PullPolicy     string
	Concurrency    int
	ResultImageTag string
	Buildtemplate  string
	RegistrySecret string // Does not belong to the service, need to be deleted
	Env            []string
	EnvSecrets     []string
	Annotations    map[string]string
	Labels         []string
	BuildArgs      []string
	BuildTimeout   string
	Cronjob        struct {
		Schedule string
		Data     string
	}
}

type registryAuths struct {
	Auths registry
}

type credentials struct {
	Username string
	Password string
}

type registry map[string]credentials

func encode(data interface{}) ([]byte, error) {
	// if output == "yaml" {
	// return yaml.Marshal(data)
	// }
	return json.MarshalIndent(data, "", "    ")
}
