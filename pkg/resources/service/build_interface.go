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
	"github.com/triggermesh/tm/pkg/resources/taskrun"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Builder interface {
	Deploy(clientset *client.ConfigSet) (string, error)
	SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error
	Delete(clientset *client.ConfigSet) error
}

func NewBuilder(s *Service) Builder {
	// if file.IsLocal(s.Source) {
	// 	return &build.Build{
	// 		Args:           s.BuildArgs,
	// 		Buildtemplate:  s.Runtime,
	// 		GenerateName:   s.Name + "-",
	// 		Namespace:      s.Namespace,
	// 		Registry:       s.Registry,
	// 		RegistrySecret: s.RegistrySecret,
	// 		Source:         s.Source,
	// 		Timeout:        s.BuildTimeout,
	// 		Wait:           true,
	// 	}
	// } else
	if file.IsLocal(s.Source) || file.IsGit(s.Source) {
		return &taskrun.TaskRun{
			Name:           s.Name,
			Namespace:      s.Namespace,
			Params:         s.BuildArgs,
			Registry:       s.Registry,
			RegistrySecret: s.RegistrySecret,
			Source: taskrun.Git{
				URL:      s.Source,
				Revision: s.Revision,
			},
			Task: taskrun.Resource{
				Name: s.Runtime,
			},
			Timeout: s.BuildTimeout,
			Wait:    true,
		}
	}
	return nil
}
