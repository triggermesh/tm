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

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
)

func newGenerateCmd(clientset *client.ConfigSet) *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate sample project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			p.Namespace = client.Namespace
			p.Runtime = args[0]
			if err := p.Generate(clientset); err != nil {
				log.Fatal(err)
			}
		},
	}
	return generateCmd
}
