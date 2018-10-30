package deploy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mholt/archiver"
	"github.com/triggermesh/tm/pkg/client"
	"k8s.io/client-go/tools/remotecommand"
)

type Copy struct {
	Pod         string
	Container   string
	Source      string
	Destination string
}

var (
	sourceTar = "/tmp/source.tar.gz"
	command   = "tar -xvf -"
)

func (c *Copy) Upload(clientset *client.ClientSet) error {
	if err := archiver.Tar.Make(sourceTar, []string{c.Source}); err != nil {
		return err
	}

	fileReader, err := os.Open(sourceTar)
	if err != nil {
		return err
	}

	if c.Destination != "" {
		command = fmt.Sprintf("%s -C %s", command, c.Destination)
	}

	_, stderr, err := c.RemoteExec(clientset, command, fileReader)
	if err != nil {
		fmt.Printf("Remote stderr: %s\n", stderr)
		return err
	}

	// fmt.Printf("Remote stdout: %s\n", output)
	return nil
}

func (c *Copy) RemoteExec(clientset *client.ClientSet, command string, file io.Reader) (string, string, error) {
	var commandLine string
	for _, v := range strings.Fields(command) {
		commandLine = fmt.Sprintf("%s&command=%s", commandLine, v)
	}
	if c.Container != "" {
		commandLine = fmt.Sprintf("&container=%s%s", c.Container, commandLine)
	}
	stdin := "false"
	if file != nil {
		stdin = "true"
	}
	url := fmt.Sprintf("%sapi/v1/namespaces/%s/pods/%s/exec?stderr=true&stdin=%s&stdout=true%s", clientset.Core.RESTClient().Post().URL().String(), clientset.Namespace, c.Pod, stdin, commandLine)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", "", err
	}

	exec, err := remotecommand.NewSPDYExecutor(clientset.Config, "POST", req.URL)
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
