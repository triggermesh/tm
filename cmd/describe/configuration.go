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

func cmdDescribeConfiguration(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"configurations"},
		Short:   "Configuration details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if args, err = listConfigurations(clientset); err != nil {
					log.Fatalln(err)
				}
			}
			for _, v := range args {
				output, err := Configuration(v, clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func listConfigurations(clientset *client.ClientSet) ([]string, error) {
	var configurations []string
	list, err := clientset.Serving.ServingV1alpha1().Configurations(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return configurations, err
	}
	for _, v := range list.Items {
		configurations = append(configurations, v.ObjectMeta.Name)
	}
	return configurations, nil
}

func Configuration(name string, clientset *client.ClientSet) ([]byte, error) {
	configuration, err := clientset.Serving.ServingV1alpha1().Configurations(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return encode(configuration)
}
