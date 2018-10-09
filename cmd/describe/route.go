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

func cmdDescribeRoute(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "route",
		Aliases: []string{"routes"},
		Short:   "Knative service route details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if args, err = listRoutes(clientset); err != nil {
					log.Fatalln(err)
				}
			}
			for _, v := range args {
				output, err := Route(v, clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func listRoutes(clientset *client.ClientSet) ([]string, error) {
	var routes []string
	list, err := clientset.Serving.ServingV1alpha1().Routes(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return routes, err
	}
	for _, v := range list.Items {
		routes = append(routes, v.ObjectMeta.Name)
	}
	return routes, nil
}

func Route(name string, clientset *client.ClientSet) ([]byte, error) {
	routes, err := clientset.Serving.ServingV1alpha1().Routes(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return encode(routes)
}
