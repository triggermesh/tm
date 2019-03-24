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

package revision

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestList(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	revisionClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	config := &Revision{Name: "Foo", Namespace: namespace}

	_, err = config.List(&revisionClient)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	revisionClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	config := &Revision{Name: "Foo", Namespace: namespace}
	result, err := config.Get(&revisionClient)
	assert.Error(t, err)
	assert.Equal(t, "", result.Name)
}

func TestDelete(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	revisionClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	config := &Revision{Name: "Foo", Namespace: namespace}
	err = config.Delete(&revisionClient)
	assert.Error(t, err)
}
