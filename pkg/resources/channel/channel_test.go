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

package channel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestList(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "test-namespace"
	}
	channelClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	channel := &Channel{Namespace: namespace}

	_, err = channel.List(&channelClient)
	assert.NoError(t, err)
}

func TestBuild(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "test-namespace"
	}
	channelClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	testCases := []struct {
		Name        string
		Provisioner string
		ErrMSG      error
	}{
		{"foo", "bar", nil},
	}

	for _, tt := range testCases {
		channel := &Channel{
			Name:        tt.Name,
			Namespace:   namespace,
			Provisioner: tt.Provisioner,
		}

		err = channel.Deploy(&channelClient)
		if err != nil {
			assert.Error(t, err)
			continue
		}

		ch, err := channel.Get(&channelClient)
		assert.NoError(t, err)
		assert.Equal(t, tt.Name, ch.Name)

		err = channel.Deploy(&channelClient)
		if err != nil {
			assert.Error(t, err)
		}

		err = channel.Delete(&channelClient)
		assert.NoError(t, err)
	}
}
