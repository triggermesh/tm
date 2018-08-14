package cmd

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
