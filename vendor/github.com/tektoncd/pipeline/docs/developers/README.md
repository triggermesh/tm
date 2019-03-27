# Developer docs

This document is aimed at helping maintainers/developers of project understand
the complexity.

## How are resources shared between tasks?

`PipelineRun` uses PVC to share resources between tasks. PVC volume is mounted
on path `/pvc` by PipelineRun.

- If a resource in a task is declared as output then the `TaskRun` controller
  adds a step to copy each output resource to the directory path
  `/pvc/task_name/resource_name`.

- If an input resource includes `from` condition then the `TaskRun` controller
  adds a step to copy from PVC to directory path
  `/pvc/previous_task/resource_name`.

Another alternatives is to use a GCS storage bucket to share the artifacts. This
can be configured using a ConfigMap with the name `config-artifact-bucket` with
the following attributes:

- location: the address of the bucket (for example gs://mybucket)
- bucket.service.account.secret.name: the name of the secret that will contain
  the credentials for the service account with access to the bucket
- bucket.service.account.secret.key: the key in the secret with the required
  service account json. The bucket is recommended to be configured with a
  retention policy after which files will be deleted.

Both options provide the same functionality to the pipeline. The choice is based
on the infrastructure used, for example in some Kubernetes platforms, the
creation of a persistent volume could be slower than uploading/downloading files
to a bucket, or if the the cluster is running in multiple zones, the access to
the persistent volume can fail.

## How are inputs handled?

Input resources, like source code (git) or artifacts, are dumped at path
`/workspace/task_resource_name`. Resource definition in task can have custom
target directory. If `targetPath` is mentioned in task input then the
controllers are responsible for adding container definitions to create
directories and also to fetch the versioned artifacts into that directory.

## How are outputs handled?

Output resources, like source code (git) or artifacts (storage resource), are
expected in directory path `/workspace/output/resource_name`.

- If resource has an output "action" like upload to blob storage, then the
  container step is added for this action.
- If there is PVC volume present (TaskRun holds owner reference to PipelineRun)
  then copy step is added as well.

- If the resource is declared only in output but not in input for task then the
  copy step includes resource being copied to PVC to path
  `/pvc/task_name/resource_name` from `/workspace/output/resource_name` like the
  following example.

  ```yaml
  kind: Task
  metadata:
    name: get-gcs-task
    namespace: default
  spec:
    outputs:
      resources:
        - name: gcs-workspace
          type: storage
  ```

- If the resource is declared both in input and output for task the then copy
  step includes resource being copied to PVC to path
  `/pvc/task_name/resource_name` from `/workspace/random-space/` if input
  resource has custom target directory (`random-space`) declared like the
  following example.

  ```yaml
  kind: Task
  metadata:
    name: get-gcs-task
    namespace: default
  spec:
    inputs:
      resources:
        - name: gcs-workspace
          type: storage
          targetPath: random-space
    outputs:
      resources:
        - name: gcs-workspace
          type: storage
  ```

  - If resource is declared both in input and output for task without custom
    target directory then copy step includes resource being copied to PVC to
    path `/pvc/task_name/resource_name` from `/workspace/random-space/` like the
    following example.

  ```yaml
  kind: Task
  metadata:
    name: get-gcs-task
    namespace: default
  spec:
    inputs:
      resources:
        - name: gcs-workspace
          type: storage
    outputs:
      resources:
        - name: gcs-workspace
          type: storage
  ```
