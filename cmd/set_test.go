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

package cmd

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type routes struct {
	Spec struct {
		Traffic []struct {
			ConfigurationName string
			RevisionName      string
			Percent           int
		}
	}
}

func TestSet(t *testing.T) {
	initConfig()

	var r routes
	name := "test-set-" + time.Now().Format("20060102150405")
	namespace = "default"
	image = "gcr.io/knative-samples/helloworld-go"

	revisions = []string{name + "-00001=50", name + "-00002=50"}

	t.Run("Describe before creation", func(t *testing.T) {
		if _, err := describeService([]string{name}); err == nil {
			t.Fatal(errors.New("Service exist before creation"))
		}
	})
	t.Run("Deploy new service", func(t *testing.T) {
		if err := deployService([]string{name}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Second)
	})
	t.Run("Adding service new revision", func(t *testing.T) {
		if err := deployService([]string{name}); err != nil {
			t.Error(err)
		}
		time.Sleep(7 * time.Second)
	})
	t.Run("Setting route traffic percentage", func(t *testing.T) {
		if err := setPercentage([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Getting route traffic percentage", func(t *testing.T) {
		data, err := describeRoute([]string{name})
		if err != nil {
			t.Error(err)
		}
		err = json.Unmarshal(data, &r)
		if err != nil {
			t.Error(err)
		}
		if len(r.Spec.Traffic) != 2 {
			t.Error(errors.New("New route not set"))
		}
		for _, v := range r.Spec.Traffic {
			if v.Percent != 50 {
				t.Error(errors.New("Incorrect traffic percentage"))
			}
		}
	})
	t.Run("Delete service", func(t *testing.T) {
		if err := deleteRoute([]string{name}); err != nil {
			t.Error(err)
		}
		if err := deleteService([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe after deletion", func(t *testing.T) {
		if _, err := describeService([]string{name}); err == nil {
			t.Fatal(errors.New("Service exist after deletion"))
		}
	})
}
