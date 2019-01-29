package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/pkg/client"
)

func TestChannelDeploy(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	channel := Channel{
		Name:        "example.com",
		Provisioner: "in-memory-channel",
	}

	err = channel.Deploy(&configSet)
	assert.NoError(t, err)
	err = delete.Channel([]string{"example.com"}, &configSet)
	assert.NoError(t, err)
}

func TestNewObject(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	channel := Channel{
		Name:        "example.com",
		Provisioner: "in-memory-channel",
	}

	newChannel := channel.newObject(&configSet)

	assert.Equal(t, channel.Name, newChannel.ObjectMeta.Name)
	assert.Equal(t, channel.Provisioner, newChannel.Spec.Provisioner.Name)
}

func TestCreateOrUpdate(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	channel := Channel{
		Name:        "testexample.com",
		Provisioner: "in-memory-channel",
	}
	newChannel := channel.newObject(&configSet)

	err = channel.createOrUpdate(newChannel, &configSet)
	assert.NoError(t, err)
	err = channel.createOrUpdate(newChannel, &configSet)
	assert.Error(t, err)
	err = delete.Channel([]string{"testexample.com"}, &configSet)
	assert.NoError(t, err)
}
