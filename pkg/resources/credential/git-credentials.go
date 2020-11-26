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

	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var secretName = "git-ssh-key"

func (g *GitCreds) CreateGitCreds(clientset *client.ConfigSet) error {
	if len(g.Key) == 0 {
		g.readStdin()
	}
	secretData := map[string]string{
		"ssh-privatekey": g.Key,
	}
	secret := corev1.Secret{
		Type: "kubernetes.io/ssh-auth",
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: g.Namespace,
			Annotations: map[string]string{
				"build.knative.dev/git-0": "*",
			},
		},
		StringData: secretData,
	}
	ctx := context.Background()
	_, err := clientset.Core.CoreV1().Secrets(client.Namespace).Create(ctx, &secret, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		oldSecret, err := clientset.Core.CoreV1().Secrets(client.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		oldSecret.StringData = secretData
		if _, err := clientset.Core.CoreV1().Secrets(client.Namespace).Update(ctx, oldSecret, metav1.UpdateOptions{}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	sa, err := clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Get(ctx, "default", metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, v := range sa.Secrets {
		if v.Name == secretName {
			return nil
		}
	}
	sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
		Name:      secretName,
		Namespace: client.Namespace,
	})
	_, err = clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (g *GitCreds) readStdin() {
	var key string
	fmt.Printf("SSH key:\n")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		key = fmt.Sprintf("%s\n%s", key, line)
	}
	g.Key = key
}
