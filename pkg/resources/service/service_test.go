// Copyright 2020 TriggerMesh Inc.
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
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

// timeout in seconds to check that resulting service is reachable
const dialTimeout = 10 * time.Second

func TestDeployAndDelete(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}

	testCases := []struct {
		name        string
		service     *Service
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid KLR function",
			service: &Service{
				Name:         "test-service-taskrun",
				Namespace:    namespace,
				Source:       "https://github.com/serverless/examples",
				Runtime:      "https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/python-3.7/runtime.yaml",
				BuildArgs:    []string{"DIRECTORY=aws-python-simple-http-endpoint", "HANDLER=handler.endpoint"},
				BuildTimeout: "15m",
			},
			wantErr: false,
		}, {
			name: "service from image",
			service: &Service{
				Name:   "test-service-from-image",
				Source: "gcr.io/google-samples/hello-app:1.0",
			},
		}, {
			name: "image not found",
			service: &Service{
				Name:   "test-missing-source",
				Source: "https://404",
			},
			wantErr:     true,
			expectedErr: "Unable to fetch image \"https://404\"",
		}, {
			name: "runtime not found",
			service: &Service{
				Name:    "test-missing-runtime",
				Source:  "https://github.com/serverless/examples",
				Runtime: "https://404",
			},
			wantErr: true,
		},
	}

	client.Dry = false
	client.Wait = true
	serviceClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.service.Namespace = namespace
			output, err := tc.service.Deploy(&serviceClient)
			if tc.wantErr {
				assert.Contains(t, err.Error(), tc.expectedErr)
				// some of the failed cases may leave kservices
				// so we'll try remove them
				tc.service.Delete(&serviceClient)
				return
			}

			address := strings.TrimPrefix(output, fmt.Sprintf("Service %s URL: https://", tc.service.Name))
			conn, err := net.DialTimeout("tcp", address+":443", dialTimeout)
			assert.NoError(t, err)
			assert.NoError(t, conn.Close())

			err = tc.service.Delete(&serviceClient)
			assert.NoError(t, err)
		})
	}
}

func TestList(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	serviceClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	s := &Service{Name: "foo", Namespace: namespace}

	_, err = s.List(&serviceClient)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}
	serviceClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	s := &Service{Name: "foo", Namespace: namespace}
	result, err := s.Get(&serviceClient)
	assert.Error(t, err)
	assert.Equal(t, "", result.Name)
}
