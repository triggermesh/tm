[![Release](https://img.shields.io/github/v/release/triggermesh/tm?label=release)](https://github.com/triggermesh/tm/releases) [![Downloads](https://img.shields.io/github/downloads/triggermesh/tm/total?label=downloads)](https://github.com/triggermesh/tm/releases) [![CircleCI](https://circleci.com/gh/triggermesh/tm/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/tm/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/triggermesh/tm)](https://goreportcard.com/report/github.com/triggermesh/tm) [![License](https://img.shields.io/github/license/triggermesh/tm?label=license)](LICENSE)

A CLI for [knative](https://github.com/knative)

## Installation

With a working [Golang](https://golang.org/doc/install) environment do:

```
go get github.com/triggermesh/tm
```

Or head to the GitHub [release page](https://github.com/triggermesh/tm/releases) and download a release.

### Configuration

**On TriggerMesh:**

1. Request beta access to TriggerMesh cloud at [https://triggermesh.com](https://triggermesh.com)
2. Download your TriggerMesh configuration file by clicking on the `download` button in the upper right corner
3. Save the file as $HOME/.tm/config.json and you are ready to use the `tm` CLI

**On your own knative cluster:**

Assuming you have access to the Kubernetes API and have a working `kubectl` setup, `tm` should work out of the box.

### Examples

Deploy service from Docker image
```
tm deploy service foo -f gcr.io/google-samples/hello-app:1.0 --wait
```

If you have Dockerfile for your service, you can use kaniko buildtemplate to deploy it
```
tm deploy service foobar \
    -f https://github.com/knative/docs \
    --build-template https://raw.githubusercontent.com/triggermesh/build-templates/master/kaniko/kaniko.yaml \
    --build-argument DIRECTORY=docs/serving/samples/hello-world/helloworld-go \
    --wait
```

or deploy service straight from Go source using Openfaas runtime
```
tm deploy service bar \
    -f https://github.com/golang/example \
    --build-template https://raw.githubusercontent.com/triggermesh/openfaas-runtime/master/go/openfaas-go-runtime.yaml \
    --build-argument DIRECTORY=hello \
    --wait
```

Moreover, for more complex deployments, tm CLI supports function definition parsing from [YAML](https://github.com/tzununbekov/serverless/blob/master/serverless.yaml) file and ability to combine multiple functions, runtimes and repositories
```
tm deploy -f https://github.com/tzununbekov/serverless
```

_If you are interested in a building image without deploying knative service, then `--build-only` flag is available in "deploy service" command_

### Running Tests Locally

To run tests you first have to set namespace you have access to with the following command:
```
export NAMESPACE=yourNamespace
```
Run unit-tests with following command from project root directory:
```
make test
```


## AWS Lambda

With triggermesh CLI you can easily deploy AWS Lambda functions on Kuberentes:

Prepare local source for Golang function

```
mkdir lambda
cd lambda
cat > main.go <<EOF
package main

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
}
EOF
```

Deploy function using Knative lambda buildtemplate with Go runtime

```
tm deploy service go-lambda -f . --build-template https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/go-1.x/buildtemplate.yaml --wait
```

Lambda function available via http events

```
curl http://go-lambda.default.dev.triggermesh.io --data '{"Name": "Foo"}'
"Hello Foo!"
```

[Here](https://github.com/triggermesh/knative-lambda-runtime) you can find more information about Knative lambda runtimes


## Deployment pipelines

_This feature is only available for Github.com repositories at the moment_

With Triggermesh CLI you can create fully functional deployment pipeline of existing git repository with a single command. In example below we're assuming that you have an access to k8s cluster with knative and tekton pipelines installed. If you use Triggermesh cloud you should not worry about requirements; platform is ready to go.

As a first step, you should create new public repository in Github.com which we will use in our example. After the empty repository has been created, we need to push sample AWS Lambda project to it:

```
tm generate python foo
cd foo
git init
git add --all
git commit -m "Sample AWS Lambda project"
git remote add origin git@github.com:<USERNAME>/<REPOSITORY>.git
git push -u origin master
```

Now that we have repository with Python project, let's create build pipeline:

```
tm push | kubectl apply -f -
```
-this command creates several knative and tekton components:

1. Tekton task with `tm` image to build AWS Lambda project using [KLR](https://github.com/triggermesh/knative-lambda-runtime)
1. Tekton taskrun to initiate project build and corresponding pipelineresource with source URL
1. Triggermesh Github custom "third-party" containersource that allows to track events on Github repositories*
1. Triggermesh Aktion [transceiver](https://github.com/triggermesh/aktion/tree/master/cmd/transceiver) and its configmap to create new taskruns on incoming events from Github containersource


\* our Github containersource is aimed at simplifying event tracking and based on periodic Github API requests (one request per minute). As a result, you don't need to create and store any tokens. Downside of this approach is that containersource have requests rate limitation (60 requests per hour) and it doesn't work with private repositories. Both of these limitation can be bypassed by providing Github personal access token in push command parameter: `tm push --token <TOKEN>`

After few minutes you should be able to see new Knative service deployed in cluster. Any commits will trigger new build and deploy so that new function will reflect all code changes.

### Docker registry

Docker images are used to run functions code in Knative services. This means that image registry is important part of service deployment scheme. Depending on type of service, Knative controller may either only pull or also push service image from and to registry. Triggermesh CLI provides simple configuration interface to setup registry address and user access credentials.

#### Service from pre-build image

Most simple type of service deployment uses service based on pre-built Docker image available in **public** registry. This kind of service doesn't require any additional configuration and may be started with following command:

```
tm deploy service foo -f gcr.io/google-samples/hello-app:1.0 --wait
```

This case doesn't produce any images so no authentication is required.

If pre-built image stored in **private** registry, you must specify access credentials by running following command before starting deployment:

```
tm set registry-auth foo-registry
```

You will be asked to enter registry address, username and password - they will saved to k8s secret and used to pull images deployed under you service account.

Besides pulling, this secret may be used to push new images for service deployment based on function source code and build template. Name of one particular k8s secret should be passed to deployment command to make CLI work with private registry:

```
tm deploy service foo-private -f https://github.com/serverless/examples \
                              --build-template knative-node4-runtime \
                              --build-argument DIRECTORY=aws-node-serve-dynamic-html-via-http-endpoint \
                              --build-argument HANDLER=handler.landingPage \
                              --registry-secret foo-registry \
                              --wait
```

If user whose credentials are specified in `foo-registry` have "write" permissions, resulting service image will be pushed to URL composed as `registry/username/service_name`

#### Gitlab CI registry

Triggermesh CLI can be used as deployment step in Gitlab CI pipeline, but considering [tokens](https://docs.gitlab.com/ee/user/project/deploy_tokens/) security policy, user must manually create CI deployment token as described [here](https://docs.gitlab.com/ee/user/project/deploy_tokens/#gitlab-deploy-token).
Deployment token must have registry read permission and should be valid for as long as the service expected to be active. If token is created, `tm` deployment step must include following commands:

```
...
script:
  - tm -n "$KUBE_NAMESPACE" set registry-auth gitlab-registry --registry "$CI_REGISTRY" --username "$CI_REGISTRY_USER" --password "$CI_JOB_TOKEN" --push
  - tm -n "$KUBE_NAMESPACE" set registry-auth gitlab-registry --registry "$CI_REGISTRY" --username "$CI_DEPLOY_USER" --password "$CI_DEPLOY_PASSWORD" --pull
...
```
After this, you may pass `--registry-secret gitlab-registry` parameter to `tm deploy` command (or in [serverless.yml](https://gitlab.com/knative-examples/functions/blob/master/serverless.yaml#L6)) so that Knative could authenticate against Gitlab registry.
Gitlab registry doesn't provide permanent read-write token that can be used in CI, but it has job-specific `CI_JOB_TOKEN` with "write" permission which is valid only while CI job running and `CI_DEPLOY_PASSWORD` with read permission which we created before. Considering this, we can see that CLI `set registry-auth` command supports `--push` and `--pull` flags that indicates which secret must be used to push image and which for "pull" operations only. Resulting images will be stored under `registry.gitlab.com/username/project/function_name` path

#### Unauthenticated registry

Besides hosted registries, triggermesh CLI may work with unauthenticated registries which does not require setting access credentials. For such cases, you may simply add `--registry-host` argument to deployment command with registry domain name parameter and resulting image will be pushed to `registry-host/namespace/service_name` URL

### Support

We would love your feedback on this CLI tool so don't hesitate to let us know what is wrong and how we could improve it, just file an [issue](https://github.com/triggermesh/tm/issues/new)

### Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)
