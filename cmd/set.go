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
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource parameters",
}

func newSetCmd(clientset *client.ConfigSet) *cobra.Command {
	setCmd.AddCommand(cmdSetRegistryCreds(clientset))
	setCmd.AddCommand(cmdSetGitCreds(clientset))
	return setCmd
}

func cmdSetRegistryCreds(clientset *client.ConfigSet) *cobra.Command {
	setRegistryCredsCmd := &cobra.Command{
		Use:   "registry-auth",
		Short: "Create secret with registry credentials",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rc.Name = args[0]
			rc.Namespace = client.Namespace
			if err := rc.CreateRegistryCreds(clientset); err != nil {
				log.Fatalln(err)
			}
			clientset.Log.Infoln("Registry credentials set")
		},
	}

	setRegistryCredsCmd.Flags().StringVar(&rc.Host, "registry", "", "Registry host address")
	setRegistryCredsCmd.Flags().StringVar(&rc.ProjectID, "project", "", "If set, use this value instead of the username in image names. Example: gcr.io/<project>")
	setRegistryCredsCmd.Flags().StringVar(&rc.Username, "username", "", "Registry username")
	setRegistryCredsCmd.Flags().StringVar(&rc.Password, "password", "", "Registry password")
	setRegistryCredsCmd.Flags().BoolVar(&rc.Pull, "pull", false, "Indicates if this token must be used for pull operations only")
	setRegistryCredsCmd.Flags().BoolVar(&rc.Push, "push", false, "Indicates if this token must be used for push operations only")
	return setRegistryCredsCmd
}

func cmdSetGitCreds(clientset *client.ConfigSet) *cobra.Command {
	setGitCredsCmd := &cobra.Command{
		Use:   "git-auth",
		Short: "Create secret with git credentials",
		Run: func(cmd *cobra.Command, args []string) {
			gc.Namespace = client.Namespace
			if err := gc.CreateGitCreds(clientset); err != nil {
				log.Fatalln(err)
			}
			clientset.Log.Infoln("Git credentials created")
		},
	}

	// Use one ssh key for all git sources for now
	// setGitCredsCmd.Flags().StringVar(&g.Host, "host", "", "Git host address")
	// setGitCredsCmd.Flags().StringVar(&g.Key, "key", "", "SSH private key")
	return setGitCredsCmd
}
