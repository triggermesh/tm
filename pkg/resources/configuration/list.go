package configuration

import (
	"encoding/json"

	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Configuration describes knative configuration object
func (c *Configuration) List(clientset *client.ConfigSet) ([]byte, error) {
	configuration, err := clientset.Serving.ServingV1alpha1().Configurations(c.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(configuration)
}
