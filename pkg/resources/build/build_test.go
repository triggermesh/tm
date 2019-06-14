// Copyright 2018 TriggerMesh, Inc
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

package build

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
	buildClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	build := &Build{Namespace: namespace}

	_, err = build.List(&buildClient)
	assert.NoError(t, err)
}

func TestBuild(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	buildClient, err := client.NewClient(client.ConfigPath(""))
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
