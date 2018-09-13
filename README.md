[![CircleCI](https://circleci.com/gh/triggermesh/tm/tree/master.svg?style=shield)](https://circleci.com/gh/triggermesh/tm/tree/master)

A cli for https://github.com/knative.build

## Install

Simply do:

```
go get github.com/triggermesh/tm
```

Or head to the GitHub release page and download it.

### Configuration

1. Register to https://frontend.munu.io
2. Press "Download tm config" button in the upper right corner
3. Save file as $HOME/.tm/config and you are read to use tm CLI

Examples:

```
tm deploy foo --from-image=gcr.io/google-samples/echo-python
```

