package cmd

import (
	"testing"
)

func TestDescribe(t *testing.T) {
	initConfig()
	namespace = "default"
	_, err := describe([]string{"testservice"})
	if err == nil {
		t.Error(err)
	}
}
