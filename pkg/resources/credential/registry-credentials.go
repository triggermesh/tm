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

package credential

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/triggermesh/tm/pkg/client"
	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetRegistryCreds creates Secret with docker registry credentials json which later can be mounted as config.json file
func (c *RegistryCreds) CreateRegistryCreds(clientset *client.ConfigSet) error {
	secrets := make(map[string]string)
	if !gitlabCI() && (len(c.Password) == 0 || len(c.Host) == 0 || len(c.Username) == 0) {
		if err := c.readStdin(); err != nil {
			return err
		}
	}
	secret := fmt.Sprintf("{\"project\":%q,\"auths\":{%q:{\"username\":%q,\"password\":%q}}}", c.ProjectID, c.Host, c.Username, c.Password)
	ctx := context.Background()
	if s, err := clientset.Core.CoreV1().Secrets(c.Namespace).Get(ctx, c.Name, metav1.GetOptions{}); err == nil {
		for k, v := range s.Data {
			secrets[k] = string(v)
		}
		if err = clientset.Core.CoreV1().Secrets(c.Namespace).Delete(ctx, c.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	if c.Pull || c.Pull == c.Push {
		secrets[".dockerconfigjson"] = secret
	}
	if c.Push || c.Push == c.Pull {
		secrets["config.json"] = secret
	}
	if _, ok := secrets[".dockerconfigjson"]; !ok {
		secrets[".dockerconfigjson"] = "{}"
	}
	newSecret := corev1.Secret{
		Type: "kubernetes.io/dockerconfigjson",
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: client.Namespace,
		},
		StringData: secrets,
	}
	if _, err := clientset.Core.CoreV1().Secrets(client.Namespace).Create(ctx, &newSecret, metav1.CreateOptions{}); err != nil {
		return err
	}

	if c.Pull || c.Pull == c.Push {
		sa, err := clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			return err
		}
		sa.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: c.Name},
		}
		if _, err := clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Update(ctx, sa, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *RegistryCreds) readStdin() error {
	reader := bufio.NewReader(os.Stdin)
	if len(c.Host) == 0 {
		fmt.Printf("Registry: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		c.Host = strings.Replace(text, "\n", "", -1)
	}
	if len(c.Username) == 0 {
		fmt.Print("Username: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		c.Username = strings.Replace(text, "\n", "", -1)
	}
	if len(c.Password) == 0 {
		fmt.Print("Password: ")
		text, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		c.Password = string(text)
	}
	return nil
}

func gitlabCI() bool {
	if ci, _ := os.LookupEnv("GITLAB_CI"); ci == "true" {
		return true
	}
	return false
}
