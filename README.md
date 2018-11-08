[![CircleCI](https://circleci.com/gh/triggermesh/tm/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/tm/tree/master)

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

Examples:

```
tm deploy service foo --from-image=gcr.io/google-samples/hello-app:1.0
```

**On your own knative cluster:**

Assuming you have access to the Kubernetes API and have a working `kubectl` setup, `tm` should work out of the box.

### Support

We would love your feedback on this CLI tool so don't hesitate to let us know what is wrong and how we could improve it, just file an [issue](https://github.com/triggermesh/tm/issues/new)

### Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)
