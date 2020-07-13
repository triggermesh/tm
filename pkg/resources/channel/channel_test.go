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

	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	channelClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	channel := &Channel{Namespace: namespace}

	_, err = channel.List(&channelClient)
	assert.NoError(t, err)
}

func TestBuild(t *testing.T) {

	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	channelClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	testCases := []struct {
		Name        string
		ExpectedErr string
	}{
		{
			Name: "foo",
		},
	}

	for _, tc := range testCases {
		channel := &Channel{
			Name:      tc.Name,
			Namespace: namespace,
		}

		err = channel.Deploy(&channelClient)
		if err != nil {
			if tc.ExpectedErr != "" {
				assert.EqualError(t, err, tc.ExpectedErr)
				continue
			}
			t.Error(err)
		}

		ch, err := channel.Get(&channelClient)
		assert.NoError(t, err)
		assert.Equal(t, tc.Name, ch.Name)

		err = channel.Delete(&channelClient)
		assert.NoError(t, err)
	}
}
