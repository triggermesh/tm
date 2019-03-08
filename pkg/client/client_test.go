/*
Copyright (c) 2018 TriggerMesh, Inc
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsername(t *testing.T) {
	assert := assert.New(t)

	c := []byte(`{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"==","server":""},"name":"test"}],"contexts":[{"context":{"cluster":"triggermesh","namespace":"testnamespace","user":"testuser"},"name":"default-context"}],"current-context":"default-context","kind":"Config","preferences":{},"users":[{"name":"testuser","user":{"token":""}}]}`)
	d := []byte(`{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"==","server":""},"name":"test"}],"contexts":[{"context":{"cluster":"test","namespace":"default","user":"testuser"},"name":"default-context"}],"current-context":"default-context","kind":"Config","preferences":{},"users":[{"name":"testuser","user":{"token":""}}]}`)

	ioutil.WriteFile("config.json", c, 0644)
	ioutil.WriteFile("default.json", d, 0644)

	testCases := []struct {
		input  string
		output string
	}{
		{"", "default"},
		{"random.json", "default"},
		{"config.json", "testnamespace"},
		{"default.json", "default"},
	}

	for _, tc := range testCases {
		namespace := getNamespace(tc.input)
		assert.Equal(tc.output, namespace)
	}

	os.Remove("config.json")
	os.Remove("default.json")
}

func TestNewClient(t *testing.T) {
	_, err := NewClient("../../testfiles/cfgfile-test.json")
	assert.NoError(t, err)

	_, err = NewClient("")
	assert.Error(t, err)
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath("")
	home := os.Getenv("HOME")
	assert.Equal(t, home+"/.tm/config.json", path)

	path = ConfigPath("../../testfiles/cfgfile-test.json")
	assert.Equal(t, "../../testfiles/cfgfile-test.json", path)

	os.Setenv("KUBECONFIG", "../../testfiles/cfgfile-test.json")
	path = ConfigPath("")
	assert.Equal(t, home+"/.tm/config.json", path)
	os.Unsetenv("KUBECONFIG")
}

func TestGetInClusterNamespace(t *testing.T) {
	namespace := getInClusterNamespace()
	assert.Equal(t, "default", namespace)
}
