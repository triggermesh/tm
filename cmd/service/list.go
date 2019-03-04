package service

import (
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(clientset *client.ConfigSet) ([]byte, error) {
	list, err := clientset.Serving.ServingV1alpha1().Services(client.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return []byte{}, err
	}
	// if output == "" {
	// table.AddRow("NAMESPACE", "SERVICE")
	// for _, item := range list.Items {
	// table.AddRow(item.Namespace, item.Name)
	// }
	// return table.String(), err
	// }
	return encode(list)
}
