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

func cmdDescribeBuild(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "Build details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if args, err = listBuilds(clientset); err != nil {
					log.Fatalln(err)
				}
			}
			for _, v := range args {
				output, err := Build(v, clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func listBuilds(clientset *client.ClientSet) ([]string, error) {
	var builds []string
	list, err := clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return builds, err
	}
	for _, v := range list.Items {
		builds = append(builds, v.ObjectMeta.Name)
	}
	return builds, nil
}

func Build(name string, clientset *client.ClientSet) ([]byte, error) {
	build, err := clientset.Build.BuildV1alpha1().Builds(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return encode(build)
}
