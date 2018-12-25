package file

import (
	"encoding/json"
	"fmt"
	"testing"
)

// ParseServerlessYAML accepts serverless yaml file path and returns decoded structure
func TestParseServerlessYAML(t *testing.T) {

	definition, err := ParseServerlessYAML("test.yaml")
	if err != nil {
		t.Errorf("ParseServerYAML failed. Expecting nil, actual %v", err)
	}

	b, err := json.MarshalIndent(definition, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}
