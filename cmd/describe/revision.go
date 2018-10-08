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

func cmdDescribeRevision(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "revision",
		Aliases: []string{"revisions"},
		Short:   "Knative revision details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				output, err := Revision(args, clientset)
				if err != nil {
					log.Errorln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

func Revision(args []string, clientset *client.ClientSet) ([]byte, error) {
	revisions, err := clientset.Serving.ServingV1alpha1().Revisions(clientset.Namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	if output == "yaml" {
		return yaml.Marshal(revisions)
	}
	return json.MarshalIndent(revisions, "", "	")
}
