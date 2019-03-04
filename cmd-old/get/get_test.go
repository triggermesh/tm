package get

import (
	"testing"

	"github.com/gosuri/uitable"
	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestNewGetCmd(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := NewGetCmd(&configSet)
	assert.Equal(t, "get", command.Use)
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"json", "json"},
		{"yaml", "yaml"},
	}

	for _, tt := range testCases {
		Format(&tt.input)
		encode(interface{}("data"))
	}

}

func TestGetBuild(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListBuild(&configSet)
	assert.Equal(t, "build", command.Use)

	output = "json"
	_, err = Builds(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Builds(&configSet)
	assert.NoError(t, err)
}

func TestGetBuildTemplates(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListBuildTemplates(&configSet)
	assert.Equal(t, "buildtemplate", command.Use)

	output = "json"
	_, err = BuildTemplates(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = BuildTemplates(&configSet)
	assert.NoError(t, err)
}

func TestGetChannels(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListChannels(&configSet)
	assert.Equal(t, "channel", command.Use)

	output = "json"
	_, err = Channels(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Channels(&configSet)
	assert.NoError(t, err)
}

func TestGetConfigurations(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListConfigurations(&configSet)
	assert.Equal(t, "configuration", command.Use)

	output = "json"
	_, err = Configurations(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Configurations(&configSet)
	assert.NoError(t, err)
}

func TestGetRevisions(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListRevision(&configSet)
	assert.Equal(t, "revision", command.Use)

	output = "json"
	_, err = Revisions(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Revisions(&configSet)
	assert.NoError(t, err)
}

func TestGetRoute(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListRoute(&configSet)
	assert.Equal(t, "route", command.Use)

	output = "json"
	_, err = Routes(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Routes(&configSet)
	assert.NoError(t, err)
}

func TestGetService(t *testing.T) {
	table = uitable.New()
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	command := cmdListService(&configSet)
	assert.Equal(t, "service", command.Use)

	output = "json"
	_, err = Services(&configSet)
	assert.NoError(t, err)

	output = ""
	_, err = Services(&configSet)
	assert.NoError(t, err)
}
