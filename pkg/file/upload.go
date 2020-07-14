/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/remotecommand"

	"github.com/triggermesh/tm/pkg/client"
)

const (
	ignoreFile        = ".tmignore"
	archivePath       = "/tmp/tm/upload"
	uploadDoneTrigger = ".uploadIsDone"
)

// Destination describes Pod destination to upload sources
type Destination struct {
	Namespace string
	Pod       string
	Container string
	Path      string
}

// NewDestination returns an instance of *Destination
// that can be used to upload sources
func NewDestination(ns, pod, container, path string) *Destination {
	return &Destination{
		Namespace: ns,
		Pod:       pod,
		Container: container,
		Path:      path,
	}
}

// Upload receives Copy structure, creates tarball of local source path and uploads it to active (un)tar process on remote pod
func (d *Destination) Upload(clientset *client.ConfigSet, basePath string) error {
	paths, err := applyIgnoreFile(basePath)
	if err != nil {
		return err
	}
	tar := path.Join(archivePath, RandString(10)+".tar")
	err = createTarball(tar, paths)
	if err != nil {
		return err
	}
	clientset.Log.Debugf("sources are packed into %q archive, opening reader\n", tar)
	fileReader, err := os.Open(tar)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	clientset.Log.Debugf("starting remote untar proccess\n")
	command := fmt.Sprintf("tar -xzvf - -C /home")
	stdout, stderr, err := d.remoteExec(clientset, command, fileReader)
	clientset.Log.Debugf("stdout:\n%s", stdout)
	clientset.Log.Debugf("stderr:\n%s", stderr)
	if err != nil {
		return err
	}
	clientset.Log.Debugf("creating upload completion flag\n")
	_, _, err = d.remoteExec(clientset, "touch "+uploadDoneTrigger, nil)
	return err
}

func readIgnoreFile(basePath string) ([]string, []string, error) {
	includes, excludes := []string{}, []string{}
	tmignore, err := os.Open(path.Join(basePath, ignoreFile))
	if err != nil {
		return includes, excludes, err
	}
	defer tmignore.Close()

	scanner := bufio.NewScanner(tmignore)
	for scanner.Scan() {
		pattern := scanner.Text()

		if strings.HasPrefix(pattern, "#") {
			continue
		}
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if pattern[0] == '!' {
			pattern = strings.TrimSpace(pattern[1:])
			includes = append(includes, pattern)
			continue
		}
		excludes = append(excludes, pattern)
	}
	return includes, excludes, nil
}

func applyIgnoreFile(basePath string) ([]string, error) {
	_, excludes, err := readIgnoreFile(basePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return []string{}, err
		}
	}

	var paths []string
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if basePath == path {
			return nil
		}
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		for _, pattern := range excludes {
			exclude, err := filepath.Match(pattern, relPath)
			if err != nil {
				return err
			}
			if exclude {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

// remoteExec executes command on remote pod and returns stdout and stderr output
func (d *Destination) remoteExec(clientset *client.ConfigSet, command string, file io.Reader) (string, string, error) {
	var commandLine string
	for _, v := range strings.Fields(command) {
		commandLine = fmt.Sprintf("%s&command=%s", commandLine, v)
	}
	if d.Container != "" {
		commandLine = fmt.Sprintf("&container=%s%s", d.Container, commandLine)
	}
	stdin := "false"
	if file != nil {
		stdin = "true"
	}
	// workaround to form correct URL
	urlAndParams := strings.Split(clientset.Core.RESTClient().Post().URL().String(), "?")
	url := fmt.Sprintf("%sapi/v1/namespaces/%s/pods/%s/exec?stderr=true&stdin=%s&stdout=true%s",
		urlAndParams[0], d.Namespace, d.Pod, stdin, commandLine)
	if len(urlAndParams) == 2 {
		url = fmt.Sprintf("%s&%s", url, urlAndParams[1])
	}
	clientset.Log.Debugf("remote exec request URL: %q\n", url)
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
