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

package describe

import (
	"errors"
	"testing"
	"time"
)

// TODO: verify returned by describe command data
func TestDescribe(t *testing.T) {
	initConfig()

	name := "test-describe-" + time.Now().Format("20060102150405")
	namespace = "default"
	source = "https://github.com/mchmarny/simple-app.git"

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
	t.Run("Describe Service", func(t *testing.T) {
		if _, err := describeService([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe Configuration", func(t *testing.T) {
		if _, err := describeConfiguration([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe Revision", func(t *testing.T) {
		if _, err := describeRevision([]string{name + "-00001"}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe Route", func(t *testing.T) {
		if _, err := describeRoute([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe Buildtemplate", func(t *testing.T) {
		if _, err := describeBuildTemplate([]string{"kaniko"}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Delete service", func(t *testing.T) {
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
