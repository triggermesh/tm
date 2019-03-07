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

package output

import (
	"encoding/json"

	"github.com/gosuri/uitable"
)

type Items []struct {
	// APIVersion string `json:"apiVersion"`
	Kind     string
	Metadata struct {
		Name      string
		Namespace string
	}
	Status struct {
		Domain     string
		Conditions []struct {
			Type   string
			Status string
		}
	}
}

func format(items Items) string {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = 80
	table.AddRow("NAMESPACE", "NAME", "READY")
	for _, item := range items {
		status := "True"
		for _, cond := range item.Status.Conditions {
			if cond.Status != "True" {
				status = cond.Status
				break
			}
		}
		table.AddRow(item.Metadata.Namespace, item.Metadata.Name, status, item.Status.Domain)
	}
	return table.String()
}

func List(list interface{}) (string, error) {
	var i Items
	data, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(data, &i); err != nil {
		return "", err
	}
	return format(i), nil
}
