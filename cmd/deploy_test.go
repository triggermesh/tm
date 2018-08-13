package cmd

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type response struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		RunLatest struct {
			Configuration struct {
				RevisionTemplate struct {
					Spec struct {
						Container struct {
							Image string `json:"image"`
						} `json:"container"`
					} `json:"spec"`
				} `json:"revisionTemplate"`
			} `json:"configuration"`
		} `json:"runLatest"`
	} `json:"spec"`
}

func TestDeploy(t *testing.T) {
	initConfig()

	var r response
	name := "go-test-" + time.Now().Format("20060102150405")
	namespace = "default"
	image = "gcr.io/knative-samples/helloworld-go"

	t.Run("Deploy", func(t *testing.T) {
		t.Log("Deploying test service")
		if err := deployService([]string{name}); err != nil {
			t.Error(err)
		}
		time.Sleep(3 * time.Second)
	})
	t.Run("Describe", func(t *testing.T) {
		t.Log("Retreiving service information")
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
	t.Run("Deploy", func(t *testing.T) {
		t.Log("Trying to update service")
		if err := deployService([]string{name}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Delete", func(t *testing.T) {
		t.Log("Deleting service")
		// Workaround. Need to make `tm delete` command use functions
		if err := serving.ServingV1alpha1().Services(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
			t.Error(err)
		}
	})
	t.Run("Get", func(t *testing.T) {
		if err := listServices([]string{name}); err != nil {
			t.Error(err)
		}
	})
}
