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
	"net/url"
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
		name    string
		service *Service
		wantErr bool
	}{
		{
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
			wantErr: true,
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
	client.Debug = true
	serviceClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.service.Namespace = namespace
			output, err := tc.service.Deploy(&serviceClient)
			if tc.wantErr {
				assert.Error(t, err)
				// some of the failed cases may leave kservices
				// so we'll try remove them
				tc.service.Delete(&serviceClient)
				return
			}

			address := strings.TrimPrefix(output, fmt.Sprintf("Service %s URL: ", tc.service.Name))
			addr, err := url.Parse(address)
			assert.NoError(t, err)
			switch addr.Scheme {
			case "https":
				address = addr.Host + ":443"
			case "http":
				address = addr.Host + ":80"
			default:
				t.Error("malformed service URL scheme", addr.Scheme)
				return
			}
			conn, err := net.DialTimeout("tcp", address, dialTimeout)
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
