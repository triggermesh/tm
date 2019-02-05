package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestDeployYaml(t *testing.T) {
	client.Dry = true
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	newService := Service{
		Namespace: client.Namespace,
	}

	services, err := newService.DeployYAML("../../testfiles/serverless-test.yaml", []string{"bar", "remote"}, &configSet)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(services))

	services, err = newService.DeployYAML("../../testfiles/serverless-test.yaml", []string{"bar", "put"}, &configSet)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(services))
}

func TestRemoveOrphans(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	newService := Service{
		Namespace: client.Namespace,
	}

	services, err := newService.DeployYAML("../../testfiles/serverless-test.yaml", []string{"bar", "remote"}, &configSet)
	assert.NoError(t, err)

	err = newService.removeOrphans(services, "", &configSet)
	assert.NoError(t, err)
}

func TestGetYAML(t *testing.T) {

	path, err := getYAML("../../testfiles/serverless-test.yaml")
	assert.NoError(t, err)

	path, err = getYAML(path)
	assert.NoError(t, err)

	_, err = getYAML("/nofile.yaml")
	assert.Error(t, err)

}
