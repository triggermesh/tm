package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/pkg/client"
)

func TestReadYAML(t *testing.T) {
	buildTemplate, err := readYAML("../../testfiles/buildtemplate-test.yaml")
	assert.NoError(t, err)
	assert.Equal(t, "nodejs-runtime", buildTemplate.ObjectMeta.Name)

	_, err = readYAML("randomfile.yaml")
	assert.Error(t, err)
}

func TestSetEnvConfig(t *testing.T) {
	buildTemplate, err := readYAML("../../testfiles/buildtemplate-test.yaml")
	assert.NoError(t, err)
	setEnvConfig("secret", &buildTemplate)
}

func TestAddSecretVolume(t *testing.T) {
	buildTemplate, err := readYAML("../../testfiles/buildtemplate-test.yaml")
	assert.NoError(t, err)
	addSecretVolume("secretVolume", &buildTemplate)
}

func TestCreateBuildTemplate(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	buildTemplate, err := readYAML("../../testfiles/buildtemplate-test.yaml")
	assert.NoError(t, err)

	buildTemplate.Namespace = client.Namespace

	err = createBuildTemplate(buildTemplate, &configSet)
	assert.NoError(t, err)

	fake2bt, err := readYAML("../../testfiles/buildtemplate-err2-test.yaml")
	assert.NoError(t, err)
	err = createBuildTemplate(fake2bt, &configSet)
	assert.Error(t, err)

	fake1bt, err := readYAML("../../testfiles/buildtemplate-err1-test.yaml")
	assert.NoError(t, err)
	err = createBuildTemplate(fake1bt, &configSet)
	assert.Error(t, err)

	testBuildtemplate := delete.Buildtemplate{
		Name:      "nodejs-runtime",
		Namespace: client.Namespace,
	}
	err = testBuildtemplate.DeleteBuildtemplate(&configSet)
	assert.NoError(t, err)
}

func TestGetBuildArguments(t *testing.T) {
	argSpec := getBuildArguments("testImage", []string{"testName=testValue"})
	assert.Equal(t, 2, len(argSpec))
	assert.Equal(t, "IMAGE", argSpec[0].Name)
	assert.Equal(t, "testImage", argSpec[0].Value)
}

func TestfBuildTemplateDeploy(t *testing.T) {
	bt := Buildtemplate{
		Name:           "testbuildTemplate",
		Namespace:      client.Namespace,
		File:           "testFile",
		RegistrySecret: "testSecret",
	}

	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	name, err := bt.Deploy(&configSet)
	assert.NoError(t, err)
	assert.Equal(t, "testbuildTemplate", name)

	testBuildtemplate := delete.Buildtemplate{
		Name:      "nodejs-runtime",
		Namespace: client.Namespace,
	}
	err = testBuildtemplate.DeleteBuildtemplate(&configSet)
	assert.NoError(t, err)
}
