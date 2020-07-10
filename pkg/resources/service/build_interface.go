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

package service

import (
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/clustertask"
	"github.com/triggermesh/tm/pkg/resources/task"
	"github.com/triggermesh/tm/pkg/resources/taskrun"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder interface contains image build methods which are common for both
// tekton pipelines and knative builds
type Builder interface {
	Deploy(clientset *client.ConfigSet) (string, error)
	SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error
	Delete(clientset *client.ConfigSet) error
}

// NewBuilder checks Service build method (tekton task)
// and returns corresponding builder interface
func NewBuilder(clientset *client.ConfigSet, s *Service) Builder {
	if !file.IsLocal(s.Source) && !file.IsGit(s.Source) {
		clientset.Log.Debugf("source %q is not local file nor git URL\n", s.Source)
		return nil
	}

	if task.Exist(clientset, s.Runtime) ||
		clustertask.Exist(clientset, s.Runtime) {
		return s.taskRun()
	}

	if file.IsRemote(s.Runtime) {
		clientset.Log.Debugf("runtime %q is seemed to be a remote file, downloading", s.Runtime)
		if localFile, err := file.Download(s.Runtime); err != nil {
			clientset.Log.Warnf("Warning! Cannot fetch runtime: %s\n", err)
		} else {
			s.Runtime = localFile
		}
	}

	return s.taskRun()
}

func (s *Service) taskRun() *taskrun.TaskRun {
	return &taskrun.TaskRun{
		Name:      s.Name,
		Namespace: s.Namespace,
		Params:    s.BuildArgs,
		Function: taskrun.Source{
			Path:     s.Source,
			Revision: s.Revision,
		},
		Task: taskrun.Resource{
			Name: s.Runtime,
		},
		Timeout: s.BuildTimeout,
		Wait:    true,
	}
}
