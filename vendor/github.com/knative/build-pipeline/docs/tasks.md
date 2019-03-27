# Tasks

A `Task` (or a [`ClusterTask`](#clustertask)) is a collection of sequential
steps you would want to run as part of your continuous integration flow. A task
will run inside a container on your cluster.

A `Task` declares:

- [Inputs](#inputs)
- [Outputs](#outputs)
- [Steps](#steps)

A `Task` is available within a namespace, and `ClusterTask` is available across
entire Kubernetes cluster.

---

- [ClusterTasks](#clustertask)
- [Syntax](#syntax)
  - [Steps](#steps)
  - [Inputs](#inputs)
  - [Outputs](#outputs)
  - [Controlling where resources are mounted](#controlling-where-resources-are-mounted)
  - [Volumes](#volumes)
  - [Templating](#templating)
- [Examples](#examples)

## ClusterTask

Similar to Task, but with a cluster scope.

In case of using a ClusterTask, the `TaskRef` kind should be added. The default
kind is Task which represents a namespaced Task

```yaml
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: demo-pipeline
  namespace: default
spec:
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
        kind: ClusterTask
      params: ....
```

A `Task` functions exactly like a `ClusterTask`, and as such all references to
`Task` below are also describing `ClusterTask`.

## Syntax

To define a configuration file for a `Task` resource, you can specify the
following fields:

- Required:
  - [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
    `tekton.dev/v1alpha1`.
  - [`kind`][kubernetes-overview] - Specify the `Task` resource object.
  - [`metadata`][kubernetes-overview] - Specifies data to uniquely identify the
    `Task` resource object, for example a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    your `Task` resource object. `Task` steps must be defined through either of
    the following fields:
    - [`steps`](#steps) - Specifies one or more container images that you want
      to run in your `Task`.
- Optional:
  - [`inputs`](#inputs) - Specifies parameters and
    [`PipelineResources`](resources.md) needed by your `Task`
  - [`outputs`](#outputs) - Specifies [`PipelineResources`](resources.md) needed
    by your `Task`
  - [`volumes`](#volumes) - Specifies one or more volumes that you want to make
    available to your build.

[kubernetes-overview]:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

The following example is a non-working sample where most of the possible
configuration fields are used:

```yaml
apiVersion: build.knative.dev/v1alpha1
kind: Task
metadata:
  name: example-task-name
spec:
  serviceAccountName: task-auth-example
  inputs:
    resources:
      - name: workspace
        type: git
    params:
      - name: pathToDockerFile
        description: The path to the dockerfile to build
        default: /workspace/workspace/Dockerfile
  outputs:
    resources:
      - name: builtImage
        type: image
  steps:
    - name: ubuntu-example
      image: ubuntu
      args: ["ubuntu-build-example", "SECRETS-example.md"]
    - image: gcr.io/example-builders/build-example
      args: ["echo", "${inputs.resources.params.pathToDockerFile}"]
    - name: dockerfile-pushexample
      image: gcr.io/example-builders/push-example
      args: ["push", "${outputs.resources.builtImage.url}"]
      volumeMounts:
        - name: docker-socket-example
          mountPath: /var/run/docker.sock
  volumes:
    - name: example-volume
      emptyDir: {}
```

### Steps

The `steps` field is required. You define one or more `steps` fields to define
the body of a `Task`.

Each `steps` in a `Task` must specify a container image that adheres to the
[container contract](./container-contract.md). For each of the `steps` fields,
or container images that you define:

- The container images are run and evaluated in order, starting from the top of
  the configuration file.
- Each container image runs until completion or until the first failure is
  detected.

### Inputs

A `Task` can declare the inputs it needs, which can be either or both of:

- [`parameters`](#parameters)
- [input resources](#input-resources)

#### Parameters

Tasks can declare input parameters that must be supplied to the task during a
TaskRun. Some example use-cases of this include:

- A Task that needs to know what compilation flags to use when building an
  application.
- A Task that needs to know what to name a built artifact.

Parameters name are limited to alpha-numeric characters, `-` and `_` and can
only start with alpha characters and `_`. For example, `fooIs-Bar_` is a valid
parameter name, `barIsBa$` or `0banana` are not.

##### Usage

The following example shows how Tasks can be parameterized, and these parameters
can be passed to the `Task` from a `TaskRun`.

Input parameters in the form of `${inputs.params.foo}` are replaced inside of
the [`steps`](#steps) (see also [templating](#templating)).

The following `Task` declares an input parameter called 'flags', and uses it in
the `steps.args` list.

```yaml
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: task-with-parameters
spec:
  inputs:
    params:
      - name: flags
        value: -someflag
  steps:
    - name: build
      image: my-builder
      args: ["build", "--flags=${inputs.params.flags}"]
```

The following `TaskRun` supplies a value for `flags`:

```yaml
apiVersion: tekton.dev/v1alpha1
kind: TaskRun
metadata:
  name: run-with-parameters
spec:
  taskRef:
    name: task-with-parameters
  inputs:
    params:
      - name: "flags"
        value: "foo=bar,baz=bat"
```

#### Input resources

Use input [`PipelineResources`](resources.md) field to provide your `Task` with
data or context that is needed by your `Task`.

Input resources, like source code (git) or artifacts, are dumped at path
`/workspace/task_resource_name` within a mounted
[volume](https://kubernetes.io/docs/concepts/storage/volumes/) and is available
to all [`steps`](#steps) of your `Task`. The path that the resources are mounted
at can be overridden with the `targetPath` value.

### Outputs

`Task` definitions can include inputs and outputs
[`PipelineResource`](resources.md) declarations. If specific set of resources
are only declared in output then a copy of resource to be uploaded or shared for
next Task is expected to be present under the path
`/workspace/output/resource_name/`.

```yaml
resources:
  outputs:
    name: storage-gcs
    type: gcs
steps:
  - image: objectuser/run-java-jar #https://hub.docker.com/r/objectuser/run-java-jar/
    command: [jar]
    args:
      ["-cvf", "-o", "/workspace/output/storage-gcs/", "projectname.war", "*"]
    env:
      - name: "FOO"
        value: "world"
```

**note**: if the task is relying on output resource functionality then the
containers in the task `steps` field cannot mount anything in the path
`/workspace/output`.

In the following example Task `tar-artifact` resource is used both as input and
output so input resource is downloaded into directory `customworkspace`(as
specified in [`targetPath`](#targetpath)). Step `untar` extracts tar file into
`tar-scratch-space` directory , `edit-tar` adds a new file and last step
`tar-it-up` creates new tar file and places in `/workspace/customworkspace/`
directory. After execution of the Task steps, (new) tar file in directory
`/workspace/customworkspace` will be uploaded to the bucket defined in
`tar-artifact` resource definition.

```yaml
resources:
  inputs:
    name: tar-artifact
    targetPath: customworkspace
  outputs:
    name: tar-artifact
steps:
 - name: untar
    image: ubuntu
    command: ["/bin/bash"]
    args: ['-c', 'mkdir -p /workspace/tar-scratch-space/ && tar -xvf /workspace/customworkspace/rules_docker-master.tar -C /workspace/tar-scratch-space/']
 - name: edit-tar
    image: ubuntu
    command: ["/bin/bash"]
    args: ['-c', 'echo crazy > /workspace/tar-scratch-space/rules_docker-master/crazy.txt']
 - name: tar-it-up
   image: ubuntu
   command: ["/bin/bash"]
   args: ['-c', 'cd /workspace/tar-scratch-space/ && tar -cvf /workspace/customworkspace/rules_docker-master.tar rules_docker-master']
```

### Controlling where resources are mounted

Tasks can opitionally provide `targetPath` to initialize resource in specific
directory. If `targetPath` is set then resource will be initialized under
`/workspace/targetPath`. If `targetPath` is not specified then resource will be
initialized under `/workspace`. Following example demonstrates how git input
repository could be initialized in `$GOPATH` to run tests:

```yaml
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: task-with-input
  namespace: default
spec:
  inputs:
    resources:
      - name: workspace
        type: git
        targetPath: go/src/github.com/knative/build-pipeline
  steps:
    - name: unit-tests
      image: golang
      command: ["go"]
      args:
        - "test"
        - "./..."
      workingDir: "/workspace/go/src/github.com/knative/build-pipeline"
      env:
        - name: GOPATH
          value: /workspace/go
```

### Volumes

Specifies one or more
[volumes](https://kubernetes.io/docs/concepts/storage/volumes/) that you want to
make available to your `Task`, including all the [`steps`](#steps). Add volumes
to complement the volumes that are implicitly created for
[input resources](#input-resources) and [output resources](#outputs).

For example, use volumes to accomplish one of the following common tasks:

- [Mount a Kubernetes secret](./auth.md).
- Create an `emptyDir` volume to act as a cache for use across multiple build
  steps. Consider using a persistent volume for inter-build caching.
- Mount
  [Kubernetes configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
  as volume source.
- Mount a host's Docker socket to use a `Dockerfile` for container image builds.
  **Note:** Building a container image using `docker build` on-cluster is _very
  unsafe_. Use [kaniko](https://github.com/GoogleContainerTools/kaniko) instead.
  This is used only for the purposes of demonstration.

### Templating

`Tasks` support templating using values from all [`inputs`](#inputs) and
[`outputs`](#outputs).

[`PipelineResources`](resources.md) can be referenced in a `Task` spec like
this, where `<name>` is the Resource Name and `<key>` is a one of the resource's
`params`:

```shell
${inputs.resources.<name>.<key>}
```

Or for an output resource:

```shell
${outputs.resources.<name>.<key>}
```

To access an input parameter, replace `resources` with `params`.

```shell
${inputs.params.<name>}
```

#### Templating Volumes

Task volume names and different [types of volumes](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) can be parameterized. Current support includes for widely used types of volumes like configmap, secret and PersistentVolumeClaim. Here is an [example](#using-kubernetes-configmap-as-volume-source) on how to use this in Task definitions.

## Examples

Use these code snippets to help you understand how to define your `Tasks`.

- [Example of image building and pushing](#example-task)
- [Mounting extra volumes](#using-an-extra-volume)
- [Mounting configMap as volume source](#using-kubernetes-configmap-as-volume-source)

_Tip: See the collection of simple
[examples](https://github.com/knative/build-pipeline/tree/master/examples) for
additional code samples._

### Example Task

For example, a `Task` to encapsulate a `Dockerfile` build might look something
like this:

**Note:** Building a container image using `docker build` on-cluster is _very
unsafe_. Use [kaniko](https://github.com/GoogleContainerTools/kaniko) instead.
This is used only for the purposes of demonstration.

```yaml
spec:
  inputs:
    resources:
      - name: workspace
        type: git
    params:
      # These may be overridden, but provide sensible defaults.
      - name: directory
        description: The directory containing the build context.
        default: /workspace
      - name: dockerfileName
        description: The name of the Dockerfile
        default: Dockerfile
  outputs:
    resources:
      - name: builtImage
        type: image
  steps:
    - name: dockerfile-build
      image: gcr.io/cloud-builders/docker
      workingDir: "${inputs.params.directory}"
      args:
        [
          "build",
          "--no-cache",
          "--tag",
          "${outputs.resources.image}",
          "--file",
          "${inputs.params.dockerfileName}",
          ".",
        ]
      volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock

    - name: dockerfile-push
      image: gcr.io/cloud-builders/docker
      args: ["push", "${outputs.resources.image}"]
      volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock

  # As an implementation detail, this template mounts the host's daemon socket.
  volumes:
    - name: docker-socket
      hostPath:
        path: /var/run/docker.sock
        type: Socket
```

#### Using an extra volume

Mounting multiple volumes:

```yaml
spec:
  steps:
    - image: ubuntu
      entrypoint: ["bash"]
      args: ["-c", "curl https://foo.com > /var/my-volume"]
      volumeMounts:
        - name: my-volume
          mountPath: /var/my-volume

    - image: ubuntu
      args: ["cat", "/etc/my-volume"]
      volumeMounts:
        - name: my-volume
          mountPath: /etc/my-volume

  volumes:
    - name: my-volume
      emptyDir: {}
```

#### Using Kubernetes Configmap as Volume Source

```yaml
spec:
  inputs:
    params:
      - name: CFGNAME
        description: Name of config map
      - name: volumeName
        description: Name of volume
  steps:
    - image: ubuntu
      entrypoint: ["bash"]
      args: ["-c", "cat /var/configmap/test"]
      volumeMounts:
        - name: "${inputs.params.volumeName}"
          mountPath: /var/configmap

  volumes:
    - name: "${inputs.params.volumeName}"
      configMap:
        name: "${inputs.params.CFGNAME}"
```

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
