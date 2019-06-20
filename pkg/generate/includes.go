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

type samplesTable map[string]service

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
	nodejsFunc = `'use strict';

module.exports.landingPage = (event, context, callback) => {
  let dynamicHtml = '<p>Hey Unknown!</p>';
  // check for GET params and use if available
  if (event.queryStringParameters && event.queryStringParameters.name) {
	dynamicHtml = ` + "`<p>Hey ${event.queryStringParameters.name}!</p>`" + `;
  }

  const html = ` + "`<html><style>h1 { color: #73757d; }</style><body><h1>Landing Page</h1>${dynamicHtml}</body></html>`" + `;

  const response = {
	statusCode: 200,
	headers: {
	  'Content-Type': 'text/html',
	},
	body: html,
  };

  // callback is sending HTML back
  callback(null, response);
};`
	packageJSON = `{
	"name": "aws-serve-dynamic-html-via-http-endpoint",
	"version": "1.0.0",
	"description": "Hookup an AWS API Gateway endpoint to a Lambda function to render HTML on a GET request",
	"author": "",
	"license": "MIT"
}`
)

func NewTable() *samplesTable {
	return &samplesTable{
		"python": service{
			source:     "handler.py",
			runtime:    "https://raw.githubusercontent.com/triggermesh/runtime-build-tasks/master/aws-lambda/python37-runtime.yaml",
			function:   pythonFunc,
			handler:    "handler.endpoint",
			apiGateway: true,
		},
		"go": service{
			source:   "main.go",
			runtime:  "https://raw.githubusercontent.com/triggermesh/runtime-build-tasks/master/aws-lambda/go-runtime.yaml",
			function: golangFunc,
		},
		"ruby": service{
			source:     "handler.rb",
			runtime:    "https://raw.githubusercontent.com/triggermesh/runtime-build-tasks/master/aws-lambda/ruby25-runtime.yaml",
			function:   rubyFunc,
			handler:    "handler.endpoint",
			apiGateway: true,
		},
		"node": service{
			source:     "handler.js",
			runtime:    "https://raw.githubusercontent.com/triggermesh/runtime-build-tasks/master/aws-lambda/node4-runtime.yaml",
			function:   nodejsFunc,
			handler:    "handler.landingPage",
			apiGateway: true,
			dependencies: []stuff{
				{name: "package.json", data: packageJSON},
			},
		},
	}
}
