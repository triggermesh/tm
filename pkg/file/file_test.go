package file

import (
	"testing"
)

// ParseServerlessYAML accepts serverless yaml file path and returns decoded structure
func TestParseServerlessYAML(t *testing.T) {

	definition, err := ParseServerlessYAML("test.yaml")
	if err != nil {
		t.Errorf("ParseServerYAML failed. Expecting nil, actual %v", err)
	}

	if definition.Service != "serverless-foo" {
		t.Errorf("Expecting `serverless-foo` actuall %v", definition.Service)
	}

	if definition.Description != "serverless.yaml parsing test" {
		t.Errorf("Expecting `serverless.yaml parsing test` actuall %v", definition.Description)
	}
}
