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

package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/triggermesh/tm/pkg/client"
)

// Upload receives Copy structure, creates tarball of local source path and uploads it to active (un)tar process on remote pod
func TestUpload(t *testing.T) {

	c := []byte(`{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"==","server":""},"name":"test"}],"contexts":[{"context":{"cluster":"triggermesh","namespace":"testnamespace","user":"testuser"},"name":"default-context"}],"current-context":"default-context","kind":"Config","preferences":{},"users":[{"name":"testuser","user":{"token":""}}]}`)

	path := os.Getenv("HOME") + "/.tm/test_config.json"
	ioutil.WriteFile(path, c, 0644)

	configSet, err := client.NewClient("", "testnamespace", "testregistry")
	if err != nil {
		t.Error(err)
	}

	copy := Copy{
		Pod:         "testPod",
		Container:   "testSourceContainer",
		Source:      "test.yaml",
		Destination: "/home",
	}

	err = copy.Upload(&configSet)
	if err != nil {
		t.Errorf("Upload failed. Expecting nil, actual %v", err)
	}

}

// RemoteExec executes command on remote pod and returns stdout and stderr output
func TestRemoteExec(t *testing.T) {
	err := os.Setenv("KUBECONFIG", "/tmp/test/config.json")
	if err != nil {
		t.Error(err)
	}

	c := []byte(`{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"==","server":""},"name":"test"}],"contexts":[{"context":{"cluster":"triggermesh","namespace":"testnamespace","user":"testuser"},"name":"default-context"}],"current-context":"default-context","kind":"Config","preferences":{},"users":[{"name":"testuser","user":{"token":""}}]}`)

	path := os.Getenv("KUBECONFIG")
	ioutil.WriteFile(path, c, 0644)

	configSet, err := client.NewClient("", "testnamespace", "testregistry")
	if err != nil {
		t.Error(err)
	}

	copy := Copy{
		Pod:         "testPod",
		Container:   "testSourceContainer",
		Source:      "test/filepath",
		Destination: "/home",
	}

	stdOut, stdErr, err := copy.RemoteExec(&configSet, "deploy", nil)
	if err != nil {
		t.Errorf("RemoteExec failed. Expecting nil, actual %v", err)
	}

	fmt.Println(stdOut)
	fmt.Println(stdErr)

	os.Remove(path)
}
