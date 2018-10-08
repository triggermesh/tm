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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cmdDescribeConfiguration(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "buildtemplate",
		Aliases: []string{"buildtemplates"},
		Short:   "Buildtemplate details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				output, err := Configuration(args, clientset)
				if err != nil {
					log.Errorln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func Configuration(args []string, clientset *client.ClientSet) ([]byte, error) {
	configuration, err := clientset.Serving.ServingV1alpha1().Configurations(clientset.Namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(configuration)
	}
	return json.MarshalIndent(configuration, "", "	")
}
