/*
Copyright (c) 2018 TriggerMesh, Inc

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

package set

import (
	"fmt"
	"testing"
)

type routes struct {
	Spec struct {
		Traffic []struct {
			ConfigurationName string
			RevisionName      string
			Percent           int
		}
	}
}

func TestSet(t *testing.T) {

	t.Run("Test Split Function", func(t *testing.T) {
		testCases := []struct {
			slice []string
		}{
			{[]string{"one=1", "two=2", "three=3"}},
			{[]string{"one:1", "two=2", "three=3"}},
			{[]string{"one= 1", "two=2", "three=3"}},
			{[]string{"one 1", "two 2", "three:=3"}},
		}

		for _, tc := range testCases {
			result := split(tc.slice)
			fmt.Println(result)
		}
	})
}
