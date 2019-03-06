package configuration

import (
	"encoding/json"

	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Configuration) Describe(clientset *client.ConfigSet) ([]byte, error) {
	configuration, err := clientset.Serving.ServingV1alpha1().Configurations(c.Namespace).Get(c.Name, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(configuration)
}
