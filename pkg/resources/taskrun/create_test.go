// Copyright 2019 TriggerMesh, Inc
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

package taskrun

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestTaskRunDryDeployment(t *testing.T) {
	namespace := "test-namespace"
	if ns, ok := os.LookupEnv("NAMESPACE"); ok {
		namespace = ns
	}

	client.Dry = true
	clientset, err := client.NewClient("../../../testfiles/cfgfile-test.json")
	assert.NoError(t, err)

	cases := []struct {
		taskrun TaskRun
		err     error
	}{{
		taskrun: TaskRun{
			Name:      "test-taskrun-success",
			Namespace: namespace,
			Registry:  "knative.registry.svc.cluster.local",
			Task: Resource{
				Name: "foo",
			},
		},
		err: nil,
	}, {
		taskrun: TaskRun{
			Name:      "",
			Namespace: namespace,
		},
		err: errors.New("taskrun name cannot be empty"),
	}, {
		taskrun: TaskRun{
			Name:      "test-taskrun-empty-taskname",
			Namespace: namespace,
			Task: Resource{
				Name: "",
			},
			PipelineResource: Resource{
				Name: "foo",
			},
		},
		err: errors.New("task name cannot be empty"),
	}, {
		taskrun: TaskRun{
			Name:      "test-taskrun-missing-resource",
			Namespace: namespace,
			Task: Resource{
				Name: "foo",
			},
			PipelineResource: Resource{
				Name: "not-existing-resource",
			},
			Function: Source{
				Path: "https://github.com/golang/example",
			},
			Registry: "knative.registry.svc.cluster.local",
		},
		err: errors.New("pipelineresource \"not-existing-resource\" not found"),
	}}

	for _, tr := range cases {
		output, err := tr.taskrun.Deploy(&clientset)
		t.Logf("%s\n%s", tr.taskrun.Name, output)
		assert.Equal(t, tr.err, err)
		if output != "" {
			assert.Contains(t, output, "\"kind\": \"TaskRun\"")
			assert.Contains(t, output, "\"apiVersion\": \"tekton.dev/v1alpha1\"")
			assert.Contains(t, output, "\"value\": \"knative.registry.svc.cluster.local/test-namespace/"+tr.taskrun.Name)
		}
	}
}
