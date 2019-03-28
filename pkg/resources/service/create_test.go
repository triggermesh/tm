package service

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/triggermesh/tm/pkg/client"
)

func TestDryRunDeployment(t *testing.T) {
	buffer := new(bytes.Buffer)
	Output = buffer
	defer func() { Output = os.Stdout }()

	client.Dry = true
	clientset, err := client.NewClient("../../../testfiles/cfgfile-test.json")
	require.NoError(t, err)

	service := &Service{Namespace: "my-namespace"}
	err = service.DeployYAML("../../../testfiles/serverless-simple.yaml", []string{}, 3, &clientset)
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "\"kind\": \"Service\"")
	assert.Contains(t, output, "\"apiVersion\": \"serving.knative.dev/v1alpha1\"")
	assert.Contains(t, output, "\"image\": \"../../../testfiles/bar/main.go\"")
}
