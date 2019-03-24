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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestDeployAndDelete(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	serviceClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	foo := &Service{
		Name:          "foo",
		Namespace:     namespace,
		Source:        "https://github.com/golang/example",
		Buildtemplate: "https://raw.githubusercontent.com/triggermesh/openfaas-runtime/master/go/openfaas-go-runtime.yaml",
		BuildArgs:     []string{"DIRECTORY=hello"},
	}

	_, err = foo.Deploy(&serviceClient)
	assert.NoError(t, err)
	err = foo.Delete(&serviceClient)
	assert.NoError(t, err)

	bar := &Service{
		Name:      "bar",
		Namespace: namespace,
		Source:    "gcr.io/google-samples/hello-app:1.0",
	}

	_, err = bar.Deploy(&serviceClient)
	assert.NoError(t, err)
	err = bar.Delete(&serviceClient)
	assert.NoError(t, err)

	foobar := &Service{
		Name:          "foobar",
		Namespace:     namespace,
		Source:        "https://github.com/knative/docs",
		Buildtemplate: "https://raw.githubusercontent.com/triggermesh/build-templates/master/kaniko/kaniko.yaml",
		BuildArgs:     []string{"DIRECTORY=serving/samples/helloworld-go"},
	}

	_, err = foobar.Deploy(&serviceClient)
	assert.NoError(t, err)
	err = foobar.Delete(&serviceClient)
	assert.NoError(t, err)

}

func TestList(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	serviceClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	s := &Service{Name: "foo", Namespace: namespace}

	_, err = s.List(&serviceClient)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	serviceClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	s := &Service{Name: "foo", Namespace: namespace}
	result, err := s.Get(&serviceClient)
	assert.Error(t, err)
	assert.Equal(t, "", result.Name)
}
