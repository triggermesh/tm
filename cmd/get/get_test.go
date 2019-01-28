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

package get

// import (
// 	"errors"
// 	"strings"
// 	"testing"
// 	"time"
// )

// func TestGet(t *testing.T) {
// 	initConfig()

// 	name := "test-get-" + time.Now().Format("20060102150405")
// 	namespace = "default"
// 	image = "gcr.io/knative-samples/helloworld-go"

// 	t.Run("Describe before creation", func(t *testing.T) {
// 		if _, err := describeService([]string{name}); err == nil {
// 			t.Fatal(errors.New("Service exist before creation"))
// 		}
// 	})
// 	t.Run("Deploy new service", func(t *testing.T) {
// 		if err := deployService([]string{name}); err != nil {
// 			t.Fatal(err)
// 		}
// 		time.Sleep(5 * time.Second)
// 	})
// 	t.Run("Get services", func(t *testing.T) {
// 		services, err := listServices()
// 		if err != nil {
// 			t.Error(err)
// 		} else {
// 			if !strings.Contains(services, name) {
// 				t.Error(errors.New("Service not found in list"))
// 			}
// 		}
// 	})
// 	t.Run("Get routes", func(t *testing.T) {
// 		routes, err := listRoutes()
// 		if err != nil {
// 			t.Error(err)
// 		} else {
// 			if !strings.Contains(routes, name) {
// 				t.Error(errors.New("Route not found in list"))
// 			}
// 		}
// 	})
// 	t.Run("Get revisions", func(t *testing.T) {
// 		revisions, err := listRevisions()
// 		if err != nil {
// 			t.Error(err)
// 		} else {
// 			if !strings.Contains(revisions, name+"-00001") {
// 				t.Error(errors.New("Revision not found in list"))
// 			}
// 		}
// 	})
// 	t.Run("Get configurations", func(t *testing.T) {
// 		configurations, err := listConfigurations()
// 		if err != nil {
// 			t.Error(err)
// 		} else {
// 			if !strings.Contains(configurations, name) {
// 				t.Error(errors.New("Configuration not found in list"))
// 			}
// 		}
// 	})
// 	t.Run("Delete service", func(t *testing.T) {
// 		if err := deleteService([]string{name}); err != nil {
// 			t.Error(err)
// 		}
// 	})
// 	t.Run("Describe after deletion", func(t *testing.T) {
// 		if _, err := describeService([]string{name}); err == nil {
// 			t.Fatal(errors.New("Service exist after deletion"))
// 		}
// 	})
// }
