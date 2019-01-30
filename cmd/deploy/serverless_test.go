package deploy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestDeployYaml(t *testing.T) {

	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	newService := Service{}

	services, err := newService.DeployYAML("git@github.com:anatoliyfedorenko/testserverlessyaml.git", []string{"get", "post"}, &configSet)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(services))

	services, err = newService.DeployYAML("git@github.com:anatoliyfedorenko/testserverlessyaml.git", []string{"get", "put"}, &configSet)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(services))
}

func TestRemoveOrphans(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	newService := Service{}

	services, err := newService.DeployYAML("git@github.com:anatoliyfedorenko/testserverlessyaml.git", []string{"get", "post"}, &configSet)
	assert.NoError(t, err)

	err = removeOrphans(services, "", &configSet)
	assert.NoError(t, err)
}

func TestGetYAML(t *testing.T) {
	path, err := getYAML("https://github.com/anatoliyfedorenko/testserverless")
	assert.Error(t, err)

	path, err = getYAML("git@github.com:anatoliyfedorenko/testserverlessyaml.git")
	assert.NoError(t, err)

	path, err = getYAML(path)
	assert.NoError(t, err)

	err = os.Remove(path)
	assert.NoError(t, err)

	_, err = getYAML("/nofile.yaml")
	assert.Error(t, err)

}
