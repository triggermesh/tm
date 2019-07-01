// // Copyright 2019 TriggerMesh, Inc
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.

package service

// import (
// 	"errors"
// 	"os"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/triggermesh/tm/pkg/client"
// )

// func TestCreateCronjobSource(t *testing.T) {
// 	namespace := "test-namespace"
// 	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
// 		namespace = ns
// 	}
// 	client.Dry = false
// 	serviceClient, err := client.NewClient(client.ConfigPath(""))
// 	assert.NoError(t, err)

// 	s := &Service{Name: "foo", Namespace: namespace}

// 	err = s.CreateCronjobSource(&serviceClient)
// 	assert.NoError(t, err)

// 	s = &Service{Name: "foo", Namespace: namespace}
// 	s.Cronjob.Schedule = "*/5 * * * *"
// 	s.Cronjob.Data = "foo"

// 	err = s.CreateCronjobSource(&serviceClient)
// 	assert.NoError(t, err)

// 	err = serviceClient.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Delete("foo-cronjob", nil)
// 	assert.NoError(t, err)
// }

// func TestParseRate(t *testing.T) {
// 	testCases := []struct {
// 		schedule string
// 		result   string
// 		err      error
// 	}{
// 		{"", "", nil},
// 		{"*/5 * * * *", "*/5 * * * *", nil},
// 		{"rate */5 * * * *", "", errors.New("invalid rate format")},
// 		{"rate(7m)", "", errors.New("rate must contain two parameters")},
// 		{"rate(7 m)", "*/7 * * * *", nil},
// 		{"rate(7 h)", "* */7 * * *", nil},
// 		{"rate(7 d)", "* * */7 * *", nil},
// 		{"rate(7 y)", "", errors.New("unknown value \"y\"")},
// 	}

// 	for _, tt := range testCases {
// 		res, err := parseRate(tt.schedule)
// 		assert.Equal(t, tt.err, err)
// 		assert.Equal(t, tt.result, res)
// 	}
// }
