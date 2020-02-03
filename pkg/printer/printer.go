// Copyright 2020 TriggerMesh, Inc
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

package printer

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
)

type headers []string
type rows [][]string

type Table struct {
	Headers headers
	Rows    rows
}

type Object struct {
	Fields    map[string]interface{}
	K8sObject interface{}
}

type Printer struct {
	Table *tablewriter.Table
}

func NewTablePrinter(out io.Writer) *Printer {
	table := tablewriter.NewWriter(out)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetRowSeparator("")
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("\t")
	return &Printer{
		Table: table,
	}
}

func (t *Printer) setTableHeaders(heads headers) {
	t.Table.SetHeader(heads)
	t.Table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.Table.SetHeaderLine(false)
}

func (t *Printer) PrintTable(table Table) {
	t.setTableHeaders(table.Headers)
	t.Table.AppendBulk(table.Rows)
	t.Table.Render()
}

func (t *Printer) PrintObject(object Object) {
	val := reflect.ValueOf(object.K8sObject)
	val = reflect.Indirect(val)
	if val.Kind().String() != "struct" {
		return
	}
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Field(i).Type()
		fieldName := val.Type().Field(i)

		for key, value := range object.Fields {
			printType := reflect.ValueOf(value).Type()
			// fmt.Printf("Comparing %q(%s) with %q(%s)\n", key, printType.String(), fieldName.Name, fieldType.String())
			if strings.EqualFold(fieldType.String(), printType.String()) {
				if fieldName.Name == key {
					output, err := yaml.Marshal(val.Field(i).Interface())
					if err != nil {
						continue
					}
					fmt.Printf("%s:\n%s\n", key, output)
					// Empty TypeMeta fields due to https://github.com/kubernetes/client-go/issues/308
					// Print only first occurrance of a requested object
					delete(object.Fields, key)
					break
				}
			}
		}

		switch fieldType.Kind().String() {
		case "struct":
			t.PrintObject(Object{
				K8sObject: val.Field(i).Interface(),
				Fields:    object.Fields,
			})
		case "slice":
			for j := 0; j < val.Field(i).Len(); j++ {
				t.PrintObject(Object{
					K8sObject: val.Field(i).Index(j).Interface(),
					Fields:    object.Fields,
				})
			}
		}
	}
}
