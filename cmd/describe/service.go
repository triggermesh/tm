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
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var err error

func cmdDescribeService(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "service",
		Aliases: []string{"services"},
		Short:   "Knative service configuration details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if args, err = listService(clientset); err != nil {
					log.Fatalln(err)
				}
			}
			for _, v := range args {
				output, err := Service(v, clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func listService(clientset *client.ClientSet) ([]string, error) {
	var services []string
	list, err := clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return services, err
	}
	for _, v := range list.Items {
		services = append(services, v.ObjectMeta.Name)
	}
	return services, nil
}

func Service(name string, clientset *client.ClientSet) ([]byte, error) {
	service, err := clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return encode(service)
}
