package build

import "encoding/json"

// Build structure represents knative build object
type Build struct {
	Name          string
	Namespace     string
	Source        string
	Revision      string
	Step          string
	Command       []string
	Buildtemplate string
	Args          []string
	Image         string
}

func encode(data interface{}) ([]byte, error) {
	// if output == "yaml" {
	// return yaml.Marshal(data)
	// }
	return json.MarshalIndent(data, "", "    ")
}
