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

package delete

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cmdDeleteRevision(clientset *client.ClientSet) *cobra.Command {
	return &cobra.Command{
		Use:     "revision",
		Short:   "Delete knative revision resource",
		Aliases: []string{"revisions"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := Revision(args, clientset); err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Revision is being deleted")
		},
	}
}

func Revision(args []string, clientset *client.ClientSet) error {
	return clientset.Serving.ServingV1alpha1().Revisions(clientset.Namespace).Delete(args[0], &metav1.DeleteOptions{})
}
