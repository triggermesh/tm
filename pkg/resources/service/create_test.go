// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}

	client.Dry = true
	clientset, err := client.NewClient("../../../testfiles/cfgfile-test.json")
	require.NoError(t, err)

	service := &Service{Namespace: namespace}
	err = service.DeployYAML("../../../testfiles/serverless-simple.yaml", []string{}, 3, &clientset)
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "\"kind\": \"Service\"")
	assert.Contains(t, output, "\"apiVersion\": \"serving.knative.dev/v1\"")
	assert.Contains(t, output, "\"image\": \"docker.io/hello-world\"")
}
