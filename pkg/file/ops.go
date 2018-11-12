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
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	tmpPath = "/tmp"
)

// Local return true if path is local filesystem
func IsLocal(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// IsRemote return true if path is URL
func IsRemote(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		return true
	}
	if _, err := http.Head("https://" + path); err == nil {
		return true
	}
	return false
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
	if resp, err := http.Head(url.String()); err == nil {
		if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 401 {
			return true
		}
	}
	return false
}

// IsRegistry return true if path "behaves" like URL to docker registry
func IsRegistry(path string) bool {
	url, err := url.Parse(path)
	if err != nil {
		return false
	}
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	resp, err := http.Head(url.String())
	if err == nil {
		if resp.StatusCode == 405 {
			return true
		}
	}
	return false
}

// Download receives URL and return path to saved file
func Download(url string) (string, error) {
	path := tmpPath + "/" + time.Now().Format(time.RFC850)
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
