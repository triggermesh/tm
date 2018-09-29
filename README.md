[![CircleCI](https://circleci.com/gh/triggermesh/tm/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/tm/tree/master)

A cli for https://github.com/knative/build

## Install

Simply do:

```
go get github.com/triggermesh/tm
```

Or head to the GitHub release page and download it.

### Configuration

1. Request beta access to TriggerMesh cloud at [https://triggermesh.com](https://triggermesh.com)
2. Press "Download tm config" ![image](https://user-images.githubusercontent.com/13515865/45539608-1084a380-b82c-11e8-9f1f-ef82e33d1e8a.png) button in the upper right corner
3. Save file as $HOME/.tm/config.json and you are read to use tm CLI

Examples:

```
tm deploy foo --from-image=gcr.io/google-samples/echo-python
```

## Deploy using a build template

With `--from-image`, the only cluster configuration that tm depends on is that the image can be pulled by cluster nodes - either public readable or with [Kubernetes configured](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/) for private registry access.

With `--from-source`, `--from-file` and `--from-url` tm needs to know more.
These commands require a build, and hence a build template.
See [below](#manage-build-templates-using-tm) for how to manage templates.
Given that a build template exists, builds also need:

 * The build template name
   - `--build-template` arg
 * The image registry that holds builds
   - The cluster has a registry specified in the tm [configuration](#configuration).
   - Override with `--registry-host` arg.
 * The build service account.
   - The cluster tm [configuration](#configuration) specifies a service account
     for the given namespace.
   - Using the `--namespace` argument, no service account is the default.
   - If service account is not set in configuration,
     builds will be created without `serviceAcountName`.
   - Override with `--service-account` arg.

### Nice-to-have configuration:

 * `--source-directory` to set `DIRECTORY` in the build spec (see for example github.com/triggermesh/nodejs-runtime)
   - might not need special handling though, if arbitrary build arguments are supported
   - source subdirectory support could possibly benefit from tm being able to guide users
 * `--build-argument` for example `--build-argument=DIRECTORY=./example-module`
   adds an argument to the generated build passed to the build template

## Manage build templates using tm

Use `tm buildtemplate` to manage build templates.
Templates have a unique name within the namespace.
Templates have no lifecycle and are thus updated using the `replace` verb.

Regular kubernetes API access can be used to manage templates, but tm includes a bit of guidance:

 * Validates that templates contain the `IMAGE` parameter.
 * Can deploy/update templates directly from a git repository that follows TriggerMesh _runtime_ conventions,
   such as [github.com/triggermesh/nodejs-runtime](https://github.com/triggermesh/nodejs-runtime).
 * Can configure the template for image pull registry
   [validation](https://github.com/triggermesh/knative-local-registry#accepting-a-cluster-generated-cert-during-build) (required until local registry TLS is supported out of the box).
