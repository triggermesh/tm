package credential

import (
	"bufio"
	"fmt"
	"os"

	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	_, err := clientset.Core.CoreV1().Secrets(client.Namespace).Create(&secret)
	if k8sErrors.IsAlreadyExists(err) {
		oldSecret, err := clientset.Core.CoreV1().Secrets(client.Namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		oldSecret.StringData = secretData
		if _, err := clientset.Core.CoreV1().Secrets(client.Namespace).Update(oldSecret); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	sa, err := clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Get("default", metav1.GetOptions{})
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
	_, err = clientset.Core.CoreV1().ServiceAccounts(client.Namespace).Update(sa)
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
