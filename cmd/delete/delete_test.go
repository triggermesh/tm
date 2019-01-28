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

package delete

// import (
// 	"errors"
// 	"testing"
// 	"time"
// )

// func TestDelete(t *testing.T) {
// 	initConfig()

// 	name := "test-delete-" + time.Now().Format("20060102150405")
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
// 	t.Run("Delete route", func(t *testing.T) {
// 		if err := deleteRoute([]string{name}); err != nil {
// 			t.Error(err)
// 		}
// 	})
// 	t.Run("Delete revision", func(t *testing.T) {
// 		if err := deleteRevision([]string{name + "-00001"}); err != nil {
// 			t.Error(err)
// 		}
// 	})
// 	t.Run("Delete configuration", func(t *testing.T) {
// 		if err := deleteConfiguration([]string{name}); err != nil {
// 			t.Error(err)
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
