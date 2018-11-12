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
	"os"
	"strings"
	"time"
)

const (
	tmpPath = "/tmp"
)

// Local return true if path is local filesystem
func Local(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// Remote return true if path is URL
func Remote(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	if _, err := http.Get(path); err == nil {
		return true
	}
	return false
}

// Git most likely return true if path is URL to git repository
func Git(path string) bool {
	if strings.HasSuffix(path, ".git") {
		return true
	}
	if resp, err := http.Get(path); err == nil {
		if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 401 {
			return true
		}
	}
	return false
}

// Registry return true if path "behaves" like URL to docker registry
func Registry(path string) bool {
	if resp, err := http.Get(path); err == nil {
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
