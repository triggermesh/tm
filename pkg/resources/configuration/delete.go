package configuration

import (
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Configuration) Delete(clientset *client.ConfigSet) error {
	return clientset.Serving.ServingV1alpha1().Configurations(c.Namespace).Delete(c.Name, &metav1.DeleteOptions{})
}
