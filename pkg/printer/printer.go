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
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
)

type headers []string
type rows [][]string

// Table is a structure with list headers and data rows that can be printed in PrintTable method
type Table struct {
	Headers headers
	Rows    rows
}

// Object is a structure that contain k8s object and field descriptions that should be printed
type Object struct {
	Fields    map[string]interface{}
	K8sObject interface{}
}

// Printer structure contains information needed to print objects in "tm get" command
type Printer struct {
	Format string
	Output io.Writer
	Table  *tablewriter.Table
}

// NewPrinter returns new Printer instance
func NewPrinter(out io.Writer) *Printer {
	table := tablewriter.NewWriter(out)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetRowSeparator("")
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("\t")
	return &Printer{
		Output: out,
		Table:  table,
	}
}

func (p *Printer) setTableHeaders(heads headers) {
	p.Table.SetHeader(heads)
	p.Table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	p.Table.SetHeaderLine(false)
}

// PrintTable accepts Table instance and prints it using olekukonko/tablewriter package
func (p *Printer) PrintTable(table Table) {
	p.setTableHeaders(table.Headers)
	p.Table.AppendBulk(table.Rows)
	p.Table.Render()
}

// PrintObject accepts Object instance and depending on output format encodes object and writes to Object output
func (p *Printer) PrintObject(object Object) error {
	switch p.Format {
	case "yaml":
		data, err := yaml.Marshal(object.K8sObject)
		if err != nil {
			return err
		}
		fmt.Fprintf(p.Output, "%s", data)
	case "json":
		data, err := json.MarshalIndent(object.K8sObject, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(p.Output, "%s", data)
	default:
		p.printShort(object)
	}
	return nil
}

func (p *Printer) printShort(object Object) {
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
					fmt.Fprintf(p.Output, "%s:\n%s\n", key, output)
					// Empty TypeMeta fields due to https://github.com/kubernetes/client-go/issues/308
					// Print only first occurrance of a requested object
					delete(object.Fields, key)
					break
				}
			}
		}

		switch fieldType.Kind().String() {
		case "struct":
			p.printShort(Object{
				K8sObject: val.Field(i).Interface(),
				Fields:    object.Fields,
			})
		case "slice":
			for j := 0; j < val.Field(i).Len(); j++ {
				p.printShort(Object{
					K8sObject: val.Field(i).Index(j).Interface(),
					Fields:    object.Fields,
				})
			}
		}
	}
}
