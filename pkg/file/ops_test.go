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
	"os"
	"testing"
)

func TestRandString(t *testing.T) {
	testCases := []int{1, 2, 3, 10, 0}
	for n, tc := range testCases {
		result := randString(tc)
		if len(result) != tc {
			t.Errorf("RandString failed on [%v]. Expecting len %v, actual %v", n, tc, len(result))
		}
		fmt.Println(result)
	}
}

func TestIsLocal(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{"", false},
		{"/", true},
	}
	for n, tc := range testCases {
		result := IsLocal(tc.path)
		if result != tc.result {
			t.Errorf("IsLocal failed on test [%v]. Expecting %v, actual %v", n, tc.result, result)
		}
	}
}

func TestIsRemote(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{"", false},
		{"https://", true},
		{"http://", true},
		{"git@", true},
		{"google.com", true},
		{"google", false},
	}
	for n, tc := range testCases {
		result := IsRemote(tc.path)
		if result != tc.result {
			t.Errorf("IsRemote failed on test [%v]. Expecting %v, actual %v", n, tc.result, result)
		}
	}
}

func TestIsGit(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{"git@", true}, //should not be true?
		{".git", true}, //should not be true?
		{"git@github.com:triggermesh/tm.git", true},
		{"https://github.com/triggermesh/tm.git", true},
		{"https://github.com/triggermesh/tm", true},
		{"github.com/triggermesh/tm", true},
		{"git@github.com:triggermesh/tm.git", true},
		{"https://triggermesh.com/", true}, //should not be true?
	}
	for n, tc := range testCases {
		result := IsGit(tc.path)
		if result != tc.result {
			t.Errorf("IsGit failed on test [%v]. Expecting %v, actual %v", n, tc.result, result)
		}
	}
}

// IsRegistry return true if path "behaves" like URL to docker registry
func TestIsRegistry(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{"google.com", false},
		{"https://triggermesh.com/", false},
		{"registry.hub.docker.com/test/testcase", true},
	}
	for n, tc := range testCases {
		result := IsRegistry(tc.path)
		if result != tc.result {
			t.Errorf("IsRegistry failed on test [%v]. Expecting %v, actual %v", n, tc.result, result)
		}
	}
}

// Download receives URL and return path to saved file
func TestDownload(t *testing.T) {
	path, err := Download("https://github.com/triggermesh/tm")
	if err != nil {
		t.Errorf("Clone failed. Expecting no error, actual %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Clone failed. Expecting folder at %v", path)
	}
	os.Remove(path)
}

// Clone runs `git clone` operation for specified URL and returns local path to repository root directory
func TestClone(t *testing.T) {
	path, err := Clone("https://github.com/triggermesh/tm")
	if err != nil {
		t.Errorf("Clone failed. Expecting no error, actual %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Clone failed. Expecting folder at %v", path)
	}
	os.Remove(path)
}
