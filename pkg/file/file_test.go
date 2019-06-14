package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ParseServerlessYAML accepts serverless yaml file path and returns decoded structure
func TestParseManifest(t *testing.T) {

	definition, err := ParseManifest("../../testfiles/serverless-test.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "serverless-foo", definition.Service)
	assert.Equal(t, "serverless.yaml parsing test", definition.Description)
}

func TestRandString(t *testing.T) {
	testCases := []int{1, 2, 3, 10, 0}
	for _, tc := range testCases {
		result := randString(tc)
		assert.Equal(t, tc, len(result))
	}
}

func TestIsLocal(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{"", false},
		{"/", true},
		{"../../testfiles/buildtemplate-test.yaml", true},
	}
	for _, tc := range testCases {
		result := IsLocal(tc.path)
		assert.Equal(t, tc.result, result)
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
	for _, tc := range testCases {
		result := IsRemote(tc.path)
		assert.Equal(t, tc.result, result)
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
	for _, tc := range testCases {
		result := IsGit(tc.path)
		assert.Equal(t, tc.result, result)
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
	for _, tc := range testCases {
		result := IsRegistry(tc.path)
		assert.Equal(t, tc.result, result)
	}
}

// Download receives URL and return path to saved file
// func TestDownload(t *testing.T) {
// 	path, err := Download("https://github.com/triggermesh/tm")
// 	assert.NoError(t, err)

// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		t.Errorf("Clone failed. Expecting folder at %v", path)
// 	}
// 	os.Remove(path)
// }

// Clone runs `git clone` operation for specified URL and returns local path to repository root directory
// func TestClone(t *testing.T) {
// 	path, err := Clone("https://github.com/triggermesh/tm")
// 	assert.NoError(t, err)

// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		t.Errorf("Clone failed. Expecting folder at %v", path)
// 	}
// 	os.Remove(path)
// }
