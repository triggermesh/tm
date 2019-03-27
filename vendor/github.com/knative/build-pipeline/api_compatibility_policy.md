# API compatibility policy

This document proposes a policy regarding making API updates to the CRDs in this
repo. Users should be able to build on the APIs in these projects with a clear
idea of what they can rely on and what should be considered in progress and
therefore likely to change.

For these purposes the CRDs are divided into three groups:

- [`Build` and `BuildTemplate`] - from https://github.com/knative/build
- [`TaskRun`, `Task`, and `ClusterTask`] - "more stable"
- [`PipelineRun`, `Pipeline` and `PipelineResource`] - "less stable"

The use of `alpha`, `beta` and `GA` in this document is meant to correspond
roughly to
[the kubernetes API deprecation policies](https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-flag-or-cli).

## What does compatibility mean here?

This policy is about changes to the APIs of the
[CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/),
aka the spec of the CRD objects themselves.

A backwards incompatible change would be a change that requires a user to update
existing instances of these CRDs in clusters where they are deployed (after
[automatic conversion is available](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/#webhook-conversion)
this process may become less painful).

The current process would look something like:

1. Backup the instances
2. Delete the instances
3. Deploy the new type definitions
4. Update the backups with the new spec
5. Deploy the updated backups

_This policy does not yet cover other functionality which could be considered
part of the API, but isn’t part of the CRD definition (e.g. a contract re. files
expected to be written in certain locations by a resulting pod)._

## `Build` and `BuildTemplate`

The CRD types
[`Build`](https://github.com/knative/docs/blob/master/build/builds.md) and
[`BuildTemplate`](https://github.com/knative/docs/blob/master/build/build-templates.md)
should be considered frozen at beta and only additive changes should be allowed.

Support will continue for the `Build` type for the foreseeable future,
particularly to support embedding of Build resources within
[`knative/serving`](https://github.com/knative/serving) objects.

## `TaskRun`, `Task`, and `ClusterTask`

The CRD types
[`Task`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#task),
[`ClusterTask`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#clustertask),
and
[`TaskRun`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#taskrun)
should be considered `alpha`, however these types are more stable than
`Pipeline`, `PipelineRun`, and `PipelineResource`.

### Possibly `beta` in 0.3

The status of these types will be revisited ~2 releases (i.e. 0.3) and see if
they can be promoted to `beta`.

Once these types are promoted to `beta`, any backwards incompatible changes must
be introduced in a backwards compatible manner first, with a deprecation warning
in the release notes, for at least one full release before the backward
incompatible change is made.

There are two reasons for this:

- `Task` and `TaskRun` are considered upgraded versions of `Build`, meaning that
  the APIs benefit from a significant amount of user feedback and iteration
- Going forward users should use `TaskRun` and `Task` instead of `Build` and
  `BuildTemplate`, those users should not expect the API to be changed on them
  without warning

The exception to this is that `PipelineResource` definitions can be embedded in
`TaskRuns`, and since the `PipelineResource` definitions are considered less
stable, changes to the spec of the embedded `PipelineResource` can be introduced
between releases.

## `PipelineRun`, `Pipeline` and `PipelineResource`

The CRD types
[`Pipeline`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#pipeline),
[`PipelineRun`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#pipelinerun)
and
[`PipelineResource`](https://github.com/knative/build-pipeline/blob/master/docs/Concepts.md#pipelineresources)
should be considered `alpha`, i.e. the API should be considered unstable.
Backwards incompatible changes can be introduced between releases, however they
must include a backwards incompatibility warning in the release notes.

The reason for this is not yet having enough user feedback to commit to the APIs
as they currently exist. Once significant user input has been given into the API
design, we can upgrade these CRDs to `beta`.
