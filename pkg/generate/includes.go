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

// SamplesTable is a set of service samples with its dependencies, names, fucntions, etc
type SamplesTable map[string]service

type service struct {
	source       string
	runtime      string
	function     string
	handler      string
	apiGateway   bool
	dependencies []stuff
}

type stuff struct {
	name string
	data string
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
	golangFunc = `package main

import (
		"fmt"
		"context"
		"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
		Name string
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
		return fmt.Sprintf("Hello %s!", name.Name ), nil
}

func main() {
		lambda.Start(HandleRequest)
}`
	rubyFunc = `def endpoint(event:, context:)
hash = {date: Time.new}
{ statusCode: 200, body: JSON.generate(hash) }
end`
	nodejsFunc = `async function justWait() {
  return new Promise((resolve, reject) => setTimeout(resolve, 100));
}

module.exports.sayHelloAsync = async (event) => {
  await justWait();
  return {hello: event && event.name || "Missing a name property in the event's JSON body"};
};`
)

// NewTable returns map with runtime name as key and service structure as value
func NewTable() *SamplesTable {
	return &SamplesTable{
		"python": service{
			source:     "handler.py",
			runtime:    "https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/python-3.7/runtime.yaml",
			function:   pythonFunc,
			handler:    "handler.endpoint",
			apiGateway: true,
		},
		"go": service{
			source:   "main.go",
			runtime:  "https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/go-1.x/runtime.yaml",
			function: golangFunc,
		},
		"ruby": service{
			source:     "handler.rb",
			runtime:    "https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/ruby-2.5/runtime.yaml",
			function:   rubyFunc,
			handler:    "handler.endpoint",
			apiGateway: true,
		},
		"node": service{
			source:   "handler.js",
			runtime:  "https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/node-10.x/runtime.yaml",
			function: nodejsFunc,
			handler:  "handler.sayHelloAsync",
		},
	}
}
