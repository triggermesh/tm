package cmd

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type service struct {
	Metadata struct {
		Name string
	}
	Spec struct {
		RunLatest struct {
			Configuration struct {
				RevisionTemplate struct {
					Spec struct {
						Container struct {
							Image string
						}
					}
				}
			}
		}
	}
	Status struct {
		LatestCreatedRevisionName string
	}
}

func TestDeploy(t *testing.T) {
	initConfig()

	var r service
	name := "test-deploy-" + time.Now().Format("20060102150405")
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
	t.Run("Describe new service", func(t *testing.T) {
		data, err := describeService([]string{name})
		if err != nil {
			t.Error(err)
		}
		err = json.Unmarshal(data, &r)
		if err != nil {
			t.Error(err)
		}
		if r.Metadata.Name != name || r.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image != image {
			t.Error(errors.New("Unexpected service name or image"))
		}
	})
	t.Run("Deploy service update", func(t *testing.T) {
		if err := deployService([]string{name}); err != nil {
			t.Error(err)
		}
		time.Sleep(7 * time.Second)
	})
	t.Run("Describe service update", func(t *testing.T) {
		data, err := describeService([]string{name})
		if err != nil {
			t.Error(err)
		}
		err = json.Unmarshal(data, &r)
		if err != nil {
			t.Error(err)
		}
		if r.Status.LatestCreatedRevisionName != name+"-00002" {
			t.Error(errors.New("Service update failed"))
		}
	})
	t.Run("Delete service", func(t *testing.T) {
		if err := deleteService([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Describe deleted service", func(t *testing.T) {
		if _, err := describeService([]string{name}); err == nil {
			t.Error(errors.New("Service left after deletion"))
		}
	})
}
