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

package generate

type samplesTable map[string]handler

type handler struct {
	name      string
	template  string
	function  string
	directory string
	handler   string
}

const (
	manifestName = "serverless.yaml"

	pythonFunc = `import json
import datetime
def endpoint(event, context):
   current_time = datetime.datetime.now().time()
   body = {
       "message": "Hello, the current time is " + str(current_time)
   }
   response = {
       "statusCode": 200,
       "body": json.dumps(body)
   }
return response`
	golangFunc = ``
	rubyFunc   = ``
	nodejsFunc = ``
)

func NewTable() *samplesTable {
	return &samplesTable{
		"python": handler{
			name:      "handler.py",
			template:  "https://raw.githubusercontent.com/triggermesh/runtime-build-tasks/master/aws-lambda/python37-runtime.yaml",
			function:  pythonFunc,
			directory: "aws-python-simple-http-endpoint",
			handler:   "handler.endpoint",
		},
		"golang": handler{
			name:     "main.go",
			function: golangFunc,
		},
		"ruby": handler{
			name:     "handler.rb",
			function: rubyFunc,
		},
		"node": handler{
			name:     "handler.js",
			function: nodejsFunc,
		},
	}
}
