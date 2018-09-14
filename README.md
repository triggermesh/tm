[![CircleCI](https://circleci.com/gh/triggermesh/tm/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/tm/tree/master)

A cli for https://github.com/knative/build

## Install

Simply do:

```
go get github.com/triggermesh/tm
```

Or head to the GitHub release page and download it.

### Configuration

Triggermesh CLI requires gcloud SDK to be [installed](https://cloud.google.com/sdk/install) with default application authorized to work with GCE API. If you already have it installed and initialized, you can configure tm CLI by following steps below:

1. Register to https://frontend.munu.io
2. Press "Download tm config" ![image](https://user-images.githubusercontent.com/13515865/45539608-1084a380-b82c-11e8-9f1f-ef82e33d1e8a.png) button in the upper right corner
3. Save file as $HOME/.tm/config.json and you are read to use tm CLI

Examples:

```
tm deploy foo --from-image=gcr.io/google-samples/echo-python
```

