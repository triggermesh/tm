// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revision

import (
	"encoding/json"

	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Revisions returns list of knative revision objects
func (r *Revision) List(clientset *client.ConfigSet) ([]byte, error) {
	list, err := clientset.Serving.ServingV1alpha1().Revisions(r.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return []byte{}, err
	}
	// if output == "" {
	// table.AddRow("NAMESPACE", "REVISION")
	// for _, item := range list.Items {
	// table.AddRow(item.Namespace, item.Name)
	// }
	// return table.String(), err
	// }
	return json.Marshal(list)
}
