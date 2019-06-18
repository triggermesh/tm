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
	"fmt"

	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/build"
	"github.com/triggermesh/tm/pkg/resources/buildtemplate"
	"github.com/triggermesh/tm/pkg/resources/clusterbuildtemplate"
	"github.com/triggermesh/tm/pkg/resources/clustertask"
	"github.com/triggermesh/tm/pkg/resources/task"
	"github.com/triggermesh/tm/pkg/resources/taskrun"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Builder interface {
	Deploy(clientset *client.ConfigSet) (string, error)
	SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error
	Delete(clientset *client.ConfigSet) error
}

func NewBuilder(clientset *client.ConfigSet, s *Service) Builder {
	if !file.IsLocal(s.Source) && !file.IsGit(s.Source) {
		return nil
	}

	if task.Exist(clientset, s.Runtime) ||
		clustertask.Exist(clientset, s.Runtime) {
		return s.taskRun()
	} else if buildtemplate.Exist(clientset, s.Runtime) ||
		clusterbuildtemplate.Exist(clientset, s.Runtime) {
		return s.build()
	}

	if file.IsRemote(s.Runtime) {
		if localFile, err := file.Download(s.Runtime); err != nil {
			fmt.Printf("Warning! Cannot fetch runtime: %s\n", err)
		} else {
			s.Runtime = localFile
		}
	}

	if file.IsBuildTemplate(s.Runtime) {
		return s.build()
	}
	return s.taskRun()
}

func (s *Service) taskRun() *taskrun.TaskRun {
	return &taskrun.TaskRun{
		Name:           s.Name,
		Namespace:      s.Namespace,
		Params:         s.BuildArgs,
		Registry:       s.Registry,
		RegistrySecret: s.RegistrySecret,
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

func (s *Service) build() *build.Build {
	fmt.Println("*******")
	fmt.Println("Warning! You're using deprecated knative/build component. Please use tekton/pipelines instead")
	fmt.Println("https://github.com/triggermesh/runtime-build-tasks")
	fmt.Println("*******")
	return &build.Build{
		Args:           s.BuildArgs,
		Buildtemplate:  s.Runtime,
		GenerateName:   s.Name + "-",
		Namespace:      s.Namespace,
		Registry:       s.Registry,
		RegistrySecret: s.RegistrySecret,
		Source:         s.Source,
		Timeout:        s.BuildTimeout,
		Wait:           true,
	}
}
