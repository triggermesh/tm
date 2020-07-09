// Copyright 2020 TriggerMesh Inc.
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
	"github.com/triggermesh/tm/pkg/push"
)

func newPushCmd(clientset *client.ConfigSet) *cobra.Command {
	var token string
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Read git repository in current directory and deploy triggermesh serverless.yaml using triggermesh and tekton components",
		Run: func(cmd *cobra.Command, args []string) {
			if err := push.Push(clientset, token); err != nil {
				log.Fatal(err)
			}
		},
	}
	pushCmd.Flags().StringVarP(&token, "token", "t", "", "Optional github access token to work with private repositories and bypass API requests limitation")

	return pushCmd
}
