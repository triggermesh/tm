package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestNewDeployCMD(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	cobraCommand := NewDeployCmd(&configSet)
	assert.Equal(t, "deploy", cobraCommand.Use)
}
