package cmd

import (
	"errors"
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	initConfig()

	name := "test-delete-" + time.Now().Format("20060102150405")
	namespace = "default"
	image = "gcr.io/knative-samples/helloworld-go"

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
	t.Run("Delete route", func(t *testing.T) {
		if err := deleteRoute([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Delete revision", func(t *testing.T) {
		if err := deleteRevision([]string{name + "-00001"}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Delete configuration", func(t *testing.T) {
		if err := deleteConfiguration([]string{name}); err != nil {
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
