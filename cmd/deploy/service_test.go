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

package deploy

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestServiceDeploy(t *testing.T) {

	client.Dry = true

	configSet, err := client.NewClient("")
	assert.NoError(t, err)

	newService := Service{
		Name:   "Test",
		Source: "testbuildtemplate.yaml",
	}

	services, err := newService.Deploy(&configSet)
	assert.NoError(t, err)
	logrus.Info(services)

	newServiceFromRepo := Service{
		Name:   "Test",
		Source: "git@github.com:anatoliyfedorenko/testserverlessyaml.git",
	}

	servicesGit, err := newServiceFromRepo.Deploy(&configSet)
	assert.NoError(t, err)
	logrus.Info(servicesGit)
}
