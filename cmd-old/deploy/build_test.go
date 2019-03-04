package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/cmd/delete"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

func TestBuildDeploy(t *testing.T) {
	configSet, err := client.NewClient("")
	assert.NoError(t, err)
	build := Build{
		Name:          "testbuild",
		Namespace:     client.Namespace,
		Buildtemplate: "knative-go-runtime",
		Step:          "buildStep",
	}
	err = build.Deploy(&configSet)
	assert.NoError(t, err)

	buildWithStep := Build{
		Name:      "testbuild",
		Namespace: client.Namespace,
		Step:      "buildStep",
		Image:     "gcr.io/example-builders/build-example",
	}
	err = buildWithStep.Deploy(&configSet)
	assert.NoError(t, err)

	testbuild := delete.Build{
		Name:      "testbuild",
		Namespace: client.Namespace,
	}
	testbuild.DeleteBuild(&configSet)
	assert.NoError(t, err)

	buildWithErr := Build{
		Name:      "testbuild",
		Namespace: client.Namespace,
	}
	err = buildWithErr.Deploy(&configSet)
	assert.Error(t, err)

	testbuild.DeleteBuild(&configSet)
	assert.Error(t, err)
}

func TestFromBuildTemplate(t *testing.T) {
	build := Build{}
	buildSpec := build.fromBuildtemplate("fooName", map[string]string{"foo": "bar"})

	assert.Equal(t, "fooName", buildSpec.Template.Name)
	assert.Equal(t, "foo", buildSpec.Template.Arguments[0].Name)
	assert.Equal(t, "bar", buildSpec.Template.Arguments[0].Value)
}

func TestFromBuildSteps(t *testing.T) {
	build := Build{}

	buildSpec := build.fromBuildSteps("testStep", "golang:alpine", []string{"/bin/bash"}, []string{"-c", "cat README.md"}, []corev1.Container{})

	assert.Equal(t, "testStep", buildSpec.Steps[0].Name)
	assert.Equal(t, "golang:alpine", buildSpec.Steps[0].Image)
}
