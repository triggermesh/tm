// Copyright 2020 TriggerMesh Inc.
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

package credential

type GitCreds struct {
	Namespace string
	// Host string
	Key string
}

// RegistryCreds contains docker registry credentials
type RegistryCreds struct {
	Name      string
	Namespace string
	Host      string
	ProjectID string
	Username  string
	Password  string
	Pull      bool
	Push      bool
}
