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

func cmdDescribeBuildTemplate(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates"},
		Short:   "BuildTemplate details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if args, err = listBuildTemplates(clientset); err != nil {
					log.Fatalln(err)
				}
			}
			for _, v := range args {
				output, err := BuildTemplate(v, clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func listBuildTemplates(clientset *client.ConfigSet) ([]string, error) {
	var buildtemplates []string
	list, err := clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Can't read BuildTemplates")
	}
	clusterlist, err := clientset.Build.BuildV1alpha1().ClusterBuildTemplates().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Can't read ClusterBuildTemplates")
	}
	for _, v := range list.Items {
		buildtemplates = append(buildtemplates, v.ObjectMeta.Name)
	}
	for _, v := range clusterlist.Items {
		buildtemplates = append(buildtemplates, v.ObjectMeta.Name)
	}
	return buildtemplates, nil
}

// BuildTemplate describes knative buildtemplate
func BuildTemplate(name string, clientset *client.ConfigSet) ([]byte, error) {
	buildtemplate, err := clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err == nil {
		return encode(buildtemplate)
	}
	clusterBuildtemplate, err := clientset.Build.BuildV1alpha1().ClusterBuildTemplates().Get(name, metav1.GetOptions{})
	if err == nil {
		return encode(clusterBuildtemplate)
	}
	return []byte{}, nil
}
