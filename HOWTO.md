# TriggerMesh CLI How-To Guide

- [TriggerMesh CLI How-To Guide](#triggermesh-cli-how-to-guide)
  - [Overview](#overview)
  - [Generating a Serverless Function](#generating-a-serverless-function)
  - [Deploying `tm` to Deploy a Service](#deploying-tm-to-deploy-a-service)
    - [Deploying a Function](#deploying-a-function)
    - [Deleteing a Function](#deleteing-a-function)
    - [Obtaining Details About a Function](#obtaining-details-about-a-function)
- [Serverless.yaml Configuration](#serverlessyaml-configuration)
  - [Top-Level Definition](#top-level-definition)
    - [Provider](#provider)
    - [Function](#function)

## Overview

The [TriggerMesh CLI][tm-cli] provides a quick and easy way to create and manage
serverless functions.

## Generating a Serverless Function

`tm generate --(python|go|ruby|node) <name>` will create a directory with
`<name>` or the name of the runtime environment passed as a flag, and
containing two files:
  * `handler.(js|rb|py|go)` The source file containing the serverless implementation
  * `serverless.yaml` The [serverless][serverless] manifest file

The [serverless.yaml](#serverlessyaml-configuration) contains the details required
to build the serverless functions, and the handler contains the specific code.

The generated handlers are designed for [AWS Lambda][aws-lambda], but make
use of the [Triggermesh Knative Lambda Runtime][tm-klr] which will allow the
handler to run on Knative, however there is no restriction to the types of
runtimes that can be used when combined with the [serverless.yaml](#serverlessyaml-configuration) file.

## Deploying `tm` to Deploy a Service

When given a `serverless.yaml` file, `tm` can be used to deploy and/or build a
serverless function.  In addition, `tm` can be used to delete the function once
deployed, and retrieve details from a running function.

The remaining commands will require a configuration file for the
[TriggerMesh Cloud](https://cloud.triggermesh.io) or a Kubernetes configuration
file.

In the target Kubernetes cluster, there may be some additional requirements:
  * [Knative Local Registry][knative-local-registry] to provide a Docker Registry for deploying built services
  * [Tekton Pipeline][tekton] for building the containers that will run the service

---

_NOTE_: An alternate Docker Registry can be used by passing `--registry-host` with
`tm`, or specifying the registry in the
[serverless.yaml](#serverlessyaml-configuration) file.  If credentials are
required to read or write to the registry, then run `tm set registry-auth <registry-secret-name>` to add the credentials to the cluster and use `--registry-secret` for the `tm deploy` command.

---

_NOTE_: If the [serverless.yaml](#serverlessyaml-configuration) file will point
to pre-built images, then Tekton will not be required.

---

The global parameters that can be used:
  * `--namespace` or `-n` The kubernetes namespace to perform the action in
  * `--registry-host` The alternate Docker registry host to use (defaulting to [Knative Local Registry][knative-local-registry])
  * `--registry-secret` The kubernetes secret in the namespace to use for authenticating against the Docker registry

Other flags to help with debugging or usability:
  * `-d` Debug mode to enable verbose output
  * `--dry` Print the Kubernetes manifests instead of applying them directly to the cluster
  * `--wait` Run the command and wait for the service to become available

### Deploying a Function

Deploying a new function or service can be done in one of two ways:
  * `serverless.yaml` file
  * Git repo, docker image, or source directory to upload

To deploy using the `serverless.yaml` file, then the command to run would be:

    tm deploy [-f path/to/serverless.yaml]

or (assuming the serverless.yaml file is in the same directory as the invocation):

    tm deploy

To deploy using the repo, docker, or source variant:

    tm depoly service <name> -f [repo|image|source] 

Note that `repo` points to a Git based repository such as
`https://github.com/triggermesh/tm.git,` `image` is a docker image such as
`docker.io/hello-world:latest`, and source is a path to the handler.

When using `repo` or `source`, the `--runtime` flag is required to reference a
Tekton task or a runtime recipe such as the
[Triggermesh Knative Lambda Runtime][tm-klr] or [Triggermesh OpenFaaS Runtime][tm-openfaas] to build the function.

Lastly, if using `repo`, then the `--revision <tag>` flag can be used to select
the branch, tag, or changeset from the repo.

### Deleteing a Function

To delete the functions defined with a serverless.yaml file:

    tm delete [-f path/to/serverless.yaml]

To delete the service:

    tm delete svc_name


### Obtaining Details About a Function

Details on the function can be obtained using:

    tm get service <svc_name>

# Serverless.yaml Configuration

The `serverless.yaml` file syntax follows a structure similar to the
[Serverless.com][serverless].

A sample `serverless.yaml` file would look like:
```yaml
service: go-demo-service
description: Sample knative service
provider:
  name: triggermesh
functions:
  go-function:
    source: main.go
    runtime: https://raw.githubusercontent.com/triggermesh/knative-lambda-runtime/master/go-1.x/runtime.yaml
```

## Top-Level Definition

| Name  | Type | Description |
|---|---|---|
|service|string|Name of this collection of services|
|description|string|_optional_ Human readable description of the services|
|provider|[TriggermeshProvider](#provider)| specific attributes|
|repository|string|_optional_ Git or local base of the serverless function repository|
|functions|map[string][function](#function)| pairs describing serverless functions|
|include|[]string|List of additional files containing function definitions|

Describes the attributes at the 'top' level of the `serverless.yaml` file.

### Provider

The provider will always be `triggermesh` when using the `tm` CLI. The other
attributes reflect where the service will be deployed, and the global parameters
to apply towards the defined functions.

| Name  | Type | Description |
|---|---|---|
|name|string|Name of the serverless provider being targeted, and must be `triggermesh`|
|pull-policy|string||_optional_ The container pull policy to apply for after the functions are built|
|namespace|string|_optional_ target namespace to deploy the services|
|runtime|string|_optional default runtime environment to use to build the services|
|buildtimeout|string|_optional_ Max runtime for building the functions before timing out the build process|
|environment|map[string]string|_optional_ Global dictionary of environment variables|
|env-secrets|[]string|_optional_ Global list of secrets that will be exposed as environment variables|
|annotations|map[string]string|_optional_ Dictionary of metadata annotations to apply to all services defined in this file|
|registry|string|_optional_ **deprecated** Docker registry server to push the services to when built|
|registry-secret|string|_optional_ **deprecated** secret name for authenticated access to the registry|

### Function

Define the attributes required for building and running a function.

| Name  | Type | Description |
|---|---|---|
|handler|string|_optional_ **deprecated** Analogous to _source_|
|source|string|_optional_ Source file that provides the function implementation|
|runtime|string|file or URL path to a yaml runtime definition on how to build the function as a container|
|buildargs|[]string|_optional_ Arguments to pass to the runtime definition during the function build process|
|description|string|_optional_ Human readable description of the function|
|labels|[]string|_optional_ Kubernetes labels to apply to the function at runtime|
|environment|map[string]string|_optional_ Environment name/value pairs to pass to the serverless function at runtime|
|env-secrets|[]string|_optional_ List of secrets which get expanded as environment variables during the function runtime|
|annotations|map[string]string|_optional_ Dictionary of metadata annotations to apply to the serverless function|

At a minimum, one of `source` or `handler` is required. If `source` points to a
file, then `runtime` will be required as well.


[tm-cli]: https://github.com/triggermesh/tm
[tm-klr]: https://github.com/triggermesh/knative-lambda-runtime
[tm-openfaas]: https://github.com/triggermesh/openfaas-runtime

[knative-local-registry]: https://github.com/triggermesh/knative-local-registry
[serverless]: https://github.com/serverless/serverless/tree/master/docs
[aws-lambda]: https://docs.aws.amazon.com/lambda
[tekton]: https://tekton.dev