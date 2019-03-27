package file

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifestValid(t *testing.T) {
	fixture, err := ioutil.ReadFile("../../testfiles/serverless-test.yaml")
	require.NoError(t, err)

	Aos = afero.NewMemMapFs()

	err = afero.WriteFile(Aos, "my-file.yml", fixture, 664)
	require.NoError(t, err)

	definition, err := ParseManifest("my-file.yml")
	require.NoError(t, err)

	assert.Equal(t, "serverless-foo", definition.Service)
	assert.Equal(t, ".", definition.Repository)
}

func TestParseManifestInvalid(t *testing.T) {
	Aos = afero.NewMemMapFs()

	err := afero.WriteFile(Aos, "my-file.yml", []byte("invalid-yaml"), 664)
	require.NoError(t, err)

	definition, err := ParseManifest("my-file.yml")

	assert.Contains(t, err.Error(), "yaml: unmarshal errors")
	assert.Empty(t, definition.Service)
}
