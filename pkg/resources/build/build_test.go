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
	"fmt"
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
		Buildtemplate string
		Args          []string
		ImageName     string
		ErrMSG        error
	}{
		{
			"kaniko-build",
			"https://github.com/knative/docs",
			"master",
			"https://raw.githubusercontent.com/triggermesh/build-templates/master/kaniko/kaniko.yaml",
			[]string{
				"DIRECTORY=docs/serving/samples/hello-world/helloworld-go",
				"FOO:BAR",
			},
			"",
			nil,
		},
	}

	for _, tt := range testCases {
		build := &Build{
			Wait:          true,
			Name:          tt.Name,
			Namespace:     namespace,
			Source:        tt.Source,
			Revision:      tt.Revision,
			Registry:      "knative.registry.svc.cluster.local",
			Buildtemplate: tt.Buildtemplate,
			Args:          tt.Args,
		}

		image, err := build.Deploy(&buildClient)
		assert.NoError(t, err)
		assert.Contains(t, image, build.Name)

		b, err := build.Get(&buildClient)
		assert.NoError(t, err)
		assert.Equal(t, tt.Name, b.Name)

		for k, v := range mapFromSlice(tt.Args) {
			present := false
			for _, buildArgs := range b.Spec.Template.Arguments {
				if buildArgs.Name == k && buildArgs.Value == v {
					present = true
					break
				}
			}
			if !present {
				assert.Error(t, fmt.Errorf("Build is missing passed arg %q", k))
				break
			}
		}

		err = build.Delete(&buildClient)
		assert.NoError(t, err)
	}
}
