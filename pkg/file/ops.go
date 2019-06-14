// Copyright 2018 TriggerMesh, Inc
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

package file

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	git "gopkg.in/src-d/go-git.v4"
)

const (
	tmpPath = "/tmp/tm/"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[random.Intn(len(letterBytes))]
	}
	return string(b)
}

// IsLocal return true if path is local filesystem
func IsLocal(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// IsDir return true if path is directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsRemote return true if path is URL
func IsRemote(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		return true
	}
	if _, err := http.Head("https://" + path); err != nil {
		return false
	}
	return true
}

// IsGit most likely return true if path is URL to git repository
func IsGit(path string) bool {
	if strings.HasSuffix(path, ".git") || strings.HasPrefix(path, "git@") {
		return true
	}
	url, err := url.Parse(path)
	if err != nil {
		return false
	}
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	resp, err := http.Head(url.String() + ".git")
	if err != nil {
		return false
	}
	for _, status := range []int{200, 301, 302, 401} {
		if resp.StatusCode == status {
			return true
		}
	}
	return false
}

func IsGitFile(path string) (string, string) {
	uri := strings.Split(path, "/")
	if len(uri) < 8 {
		return "", ""
	}
	repository := fmt.Sprintf("https://%s/%s/%s", uri[0], uri[1], uri[2])
	file := strings.Join(uri[5:], "/")
	if strings.HasPrefix(uri[0], "http") {
		repository = fmt.Sprintf("%s//%s/%s/%s", uri[0], uri[2], uri[3], uri[4])
		file = strings.Join(uri[7:], "/")
	}
	if !IsGit(repository) {
		return "", ""
	}
	return repository, file
}

// IsRegistry return true if path "behaves" like URL to docker registry
func IsRegistry(path string) bool {
	url, err := url.Parse(path)
	if err != nil {
		return false
	}
	// Need to find the way how to detect docker registry
	if strings.Contains(url.String(), "docker.") {
		return true
	}
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	resp, err := http.Head(url.String())
	if err != nil {
		return false
	}
	if resp.StatusCode == 405 {
		return true
	}
	return false
}

// Download receives URL and return path to saved file
func Download(url string) (string, error) {
	path := fmt.Sprintf("%s/download", tmpPath)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	path = fmt.Sprintf("%s/%s", path, RandString(10))
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}

// Clone runs `git clone` operation for specified URL and returns local path to repository root directory
func Clone(url string) (string, error) {
	path := fmt.Sprintf("%s/git/%s", tmpPath, RandString(10))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	return path, err
}
