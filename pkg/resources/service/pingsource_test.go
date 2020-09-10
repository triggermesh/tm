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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
)

func TestPingSource(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}

	testCases := []struct {
		name              string
		service           *Service
		pingSourceCreated bool
		reason            string
	}{
		{
			name: "valid pingsource",
			service: &Service{
				Name:   "valid-pingsource",
				Source: "gcr.io/google-samples/hello-app:1.0",
				Schedule: []file.Schedule{
					{
						Cron:     "*/1 * * * *",
						JSONData: `{"some":"data"}`,
					},
				},
			},
			pingSourceCreated: true,
		}, {
			name: "malformed schedule",
			service: &Service{
				Name:   "malformed-schedule",
				Source: "gcr.io/google-samples/hello-app:1.0",
				Schedule: []file.Schedule{
					{
						Cron: "not a schedule",
					},
				},
			},
			pingSourceCreated: false,
		},
	}

	client.Dry = false
	client.Wait = false
	serviceClient, err := client.NewClient(client.ConfigPath(""))
	assert.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.service.Namespace = namespace
			_, err := tc.service.Deploy(&serviceClient)
			assert.NoError(t, err)
			defer tc.service.Delete(&serviceClient)

			psList, err := serviceClient.Eventing.SourcesV1alpha2().PingSources(tc.service.Namespace).List(metav1.ListOptions{
				LabelSelector: serviceLabelKey + "=" + tc.service.Name,
			})
			assert.NoError(t, err)

			if !tc.pingSourceCreated {
				assert.Len(t, psList.Items, 0)
				return
			}

			// multiple schedules are not supported in tests
			assert.Len(t, psList.Items, 1)
			ps := psList.Items[0]

			assert.Equal(t, ps.Spec.Sink.Ref.Name, tc.service.Name)
			assert.Equal(t, ps.Spec.Sink.Ref.Namespace, tc.service.Namespace)

			assert.Len(t, ps.OwnerReferences, 1)
			assert.Equal(t, ps.OwnerReferences[0].Kind, "Service")
			assert.Equal(t, ps.OwnerReferences[0].Name, tc.service.Name)

			assert.Equal(t, ps.Spec.JsonData, tc.service.Schedule[0].JSONData)
			assert.Equal(t, ps.Spec.Schedule, tc.service.Schedule[0].Cron)
		})
	}
}
