package serverless

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type File struct {
	Service  string
	Provider struct {
		Name                 string
		Runtime              string
		DefaultDNSResolution string
		Environment          map[string]string
	}
	Plugins    []string
	Repository string
	Functions  map[string]Function
}

type Function struct {
	Handler string
	Port    int
	Events  []map[string]interface{}
}

func Parse(path string) (File, error) {
	var f File
	if _, err := os.Stat(path); err != nil {
		return f, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return f, err
	}
	f.Repository = filepath.Base(filepath.Dir(path))
	err = yaml.Unmarshal(data, f)

	return f, err
}
