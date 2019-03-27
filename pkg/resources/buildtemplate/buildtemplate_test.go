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

package buildtemplate

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestList(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	buildTemplateClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	buildtemplate := &Buildtemplate{Namespace: namespace}

	_, err = buildtemplate.List(&buildTemplateClient)
	assert.NoError(t, err)
}

func TestBuildTemplate(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	buildTemplateClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	testCases := []struct {
		Name           string
		File           string
		RegistrySecret string
		ErrMSG         error
	}{
		{"foo", "", "", errors.New("Buildtemplate : Get : unsupported protocol scheme \"\"")},
		//{"foo", "https://github.com/triggermesh/tm/blob/master/testfiles/broken-buildtemplate.yaml", "", errors.New("error converting YAML to JSON: yaml: line 526: mapping values are not allowed in this context")},
		{"foo", "../../../testfiles/buildtemplate-err1-test.yaml", "", errors.New("Build template \"IMAGE\" parameter is missing")},
		{"foo", "../../../testfiles/buildtemplate-err2-test.yaml", "", errors.New("Can't create object, only BuildTemplate is allowed")},
		{"foo", "../../../testfiles/buildtemplate-test.yaml", "", nil},
		{"foo", "../../../testfiles/buildtemplate-test.yaml", "secretBar", nil},
	}

	for _, tt := range testCases {
		buildtemplate := &Buildtemplate{
			Name:           tt.Name,
			File:           tt.File,
			RegistrySecret: tt.RegistrySecret,
			Namespace:      namespace,
		}

		_, err := buildtemplate.Deploy(&buildTemplateClient)
		if err != nil {
			assert.Equal(t, tt.ErrMSG, err)
			continue
		}

		bt, err := buildtemplate.Get(&buildTemplateClient)
		assert.NoError(t, err)
		assert.Equal(t, tt.Name, bt.Name)

		_, err = buildtemplate.Deploy(&buildTemplateClient)
		if err != nil {
			assert.Equal(t, tt.ErrMSG, err)
		}

		err = buildtemplate.Delete(&buildTemplateClient)
		assert.NoError(t, err)
	}
}
