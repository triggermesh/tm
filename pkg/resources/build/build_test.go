package build

import (
	"errors"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestList(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	buildClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)
	log.Info(buildClient)

	build := &Build{Namespace: namespace}

	buildList, err := build.List(&buildClient)
	assert.NoError(t, err)
	log.Info(buildList)
}

func TestBuild(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	buildClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	testCases := []struct {
		Name          string
		Source        string
		Revision      string
		Step          string
		Command       []string
		Buildtemplate string
		Args          []string
		Image         string
		ErrMSG        error
	}{
		{"foo", "", "", "", []string{}, "", []string{}, "", errors.New("Build steps or buildtemplate name must be specified")},
		{"foo", "", "", "", []string{}, "https://raw.githubusercontent.com/triggermesh/build-templates/master/kaniko/kaniko.yaml", []string{"DIRECTORY=serving/samples/helloworld-go", "FOO:BAR", "FOO%BAR"}, "", nil},
		{"foo", "", "", "build", []string{}, "", []string{}, "", nil},
	}

	for _, tt := range testCases {
		build := &Build{
			Name:          tt.Name,
			Namespace:     namespace,
			Source:        tt.Source,
			Revision:      tt.Revision,
			Step:          tt.Step,
			Command:       tt.Command,
			Buildtemplate: tt.Buildtemplate,
			Args:          tt.Args,
			Image:         tt.Image,
		}

		err = build.Deploy(&buildClient)
		if err != nil {
			assert.Error(t, err)
			continue
		}

		b, err := build.Get(&buildClient)
		assert.NoError(t, err)
		assert.Equal(t, tt.Name, b.Name)

		err = build.DeleteBuild(&buildClient)
		assert.NoError(t, err)
	}
}
