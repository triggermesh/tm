/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"
	"strings"
	"time"

	"github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/knative/build-pipeline/pkg/names"
	"github.com/knative/build-pipeline/pkg/reconciler/v1alpha1/taskrun/list"
	"github.com/knative/build-pipeline/pkg/reconciler/v1alpha1/taskrun/resources"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ReasonRunning indicates that the reason for the inprogress status is that the TaskRun
	// is just starting to be reconciled
	ReasonRunning = "Running"

	// ReasonFailed indicates that the reason for the failure status is that one of the TaskRuns failed
	ReasonFailed = "Failed"

	// ReasonSucceeded indicates that the reason for the finished status is that all of the TaskRuns
	// completed successfully
	ReasonSucceeded = "Succeeded"

	// ReasonTimedOut indicates that the PipelineRun has taken longer than its configured
	// timeout
	ReasonTimedOut = "PipelineRunTimeout"
)

// GetNextTask returns the next Task for which a TaskRun should be created,
// or nil if no TaskRun should be created.
func GetNextTask(prName string, state []*ResolvedPipelineRunTask, logger *zap.SugaredLogger) *ResolvedPipelineRunTask {
	for _, prtr := range state {
		if prtr.TaskRun != nil {
			c := prtr.TaskRun.Status.GetCondition(duckv1alpha1.ConditionSucceeded)
			if c == nil {
				logger.Infof("TaskRun %s doesn't have a condition so it is just starting and we shouldn't start more for PipelineRun %s", prtr.TaskRunName, prName)
				return nil
			}
			switch c.Status {
			case corev1.ConditionFalse:
				logger.Infof("TaskRun %s has failed; we don't need to run PipelineRun %s", prtr.TaskRunName, prName)
				return nil
			case corev1.ConditionUnknown:
				logger.Infof("TaskRun %s is still running so we shouldn't start more for PipelineRun %s", prtr.TaskRunName, prName)
				return nil
			}
		} else {
			logger.Infof("TaskRun %s should be started for PipelineRun %s", prtr.TaskRunName, prName)
			return prtr
		}
	}
	logger.Infof("No TaskRuns to start for PipelineRun %s", prName)
	return nil
}

// ResolvedPipelineRunTask contains a Task and its associated TaskRun, if it
// exists. TaskRun can be nil to represent there being no TaskRun.
type ResolvedPipelineRunTask struct {
	TaskRunName           string
	TaskRun               *v1alpha1.TaskRun
	PipelineTask          *v1alpha1.PipelineTask
	ResolvedTaskResources *resources.ResolvedTaskResources
}

// GetTaskRun is a function that will retrieve the TaskRun name.
type GetTaskRun func(name string) (*v1alpha1.TaskRun, error)

// GetResourcesFromBindings will validate that all PipelineResources declared in Pipeline p are bound in PipelineRun pr
// and if so, will return a map from the declared name of the PipelineResource (which is how the PipelineResource will
// be referred to in the PipelineRun) to the ResourceRef.
func GetResourcesFromBindings(p *v1alpha1.Pipeline, pr *v1alpha1.PipelineRun) (map[string]v1alpha1.PipelineResourceRef, error) {
	resources := map[string]v1alpha1.PipelineResourceRef{}

	required := make([]string, 0, len(p.Spec.Resources))
	for _, resource := range p.Spec.Resources {
		required = append(required, resource.Name)
	}
	provided := make([]string, 0, len(pr.Spec.Resources))
	for _, resource := range pr.Spec.Resources {
		provided = append(provided, resource.Name)
	}
	err := list.IsSame(required, provided)
	if err != nil {
		return resources, fmt.Errorf("PipelineRun bound resources didn't match Pipeline: %s", err)
	}

	for _, resource := range pr.Spec.Resources {
		resources[resource.Name] = resource.ResourceRef
	}
	return resources, nil
}

func getPipelineRunTaskResources(pt v1alpha1.PipelineTask, providedResources map[string]v1alpha1.PipelineResourceRef) ([]v1alpha1.TaskResourceBinding, []v1alpha1.TaskResourceBinding, error) {
	inputs, outputs := []v1alpha1.TaskResourceBinding{}, []v1alpha1.TaskResourceBinding{}
	if pt.Resources != nil {
		for _, taskInput := range pt.Resources.Inputs {
			resource, ok := providedResources[taskInput.Resource]
			if !ok {
				return inputs, outputs, fmt.Errorf("pipelineTask tried to use input resource %s not present in declared resources", taskInput.Resource)
			}
			inputs = append(inputs, v1alpha1.TaskResourceBinding{
				Name:        taskInput.Name,
				ResourceRef: resource,
			})
		}
		for _, taskOutput := range pt.Resources.Outputs {
			resource, ok := providedResources[taskOutput.Resource]
			if !ok {
				return outputs, outputs, fmt.Errorf("pipelineTask tried to use output resource %s not present in declared resources", taskOutput.Resource)
			}
			outputs = append(outputs, v1alpha1.TaskResourceBinding{
				Name:        taskOutput.Name,
				ResourceRef: resource,
			})
		}
	}
	return inputs, outputs, nil
}

// TaskNotFoundError indicates that the resolution failed because a referenced Task couldn't be retrieved
type TaskNotFoundError struct {
	Name string
	Msg  string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("Couldn't retrieve Task %q: %s", e.Name, e.Msg)
}

// ResourceNotFoundError indicates that the resolution failed because a referenced PipelineResource couldn't be retrieved
type ResourceNotFoundError struct {
	Msg string
}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("Couldn't retrieve PipelineResource: %s", e.Msg)
}

// ResolvePipelineRun retrieves all Tasks instances which are reference by tasks, getting
// instances from getTask. If it is unable to retrieve an instance of a referenced Task, it
// will return an error, otherwise it returns a list of all of the Tasks retrieved.
// It will retrieve the Resources needed for the TaskRun as well using getResource and the mapping
// of providedResources.
func ResolvePipelineRun(
	pipelineRun v1alpha1.PipelineRun,
	getTask resources.GetTask,
	getClusterTask resources.GetClusterTask,
	getResource resources.GetResource,
	tasks []v1alpha1.PipelineTask,
	providedResources map[string]v1alpha1.PipelineResourceRef,
) ([]*ResolvedPipelineRunTask, error) {

	state := []*ResolvedPipelineRunTask{}
	for i := range tasks {
		pt := tasks[i]

		rprt := ResolvedPipelineRunTask{
			PipelineTask: &pt,
			TaskRunName:  getTaskRunName(pipelineRun.Status.TaskRuns, pipelineRun.Name, &pt),
		}

		// Find the Task that this task in the Pipeline this PipelineTask is using
		var t v1alpha1.TaskInterface
		var err error
		if pt.TaskRef.Kind == v1alpha1.ClusterTaskKind {
			t, err = getClusterTask(pt.TaskRef.Name)
		} else {
			t, err = getTask(pt.TaskRef.Name)
		}
		if err != nil {
			return nil, &TaskNotFoundError{
				Name: pt.TaskRef.Name,
				Msg:  err.Error(),
			}
		}

		// Get all the resources that this task will be using, if any
		inputs, outputs, err := getPipelineRunTaskResources(pt, providedResources)
		if err != nil {
			return nil, fmt.Errorf("unexpected error which should have been caught by Pipeline webhook: %v", err)
		}

		spec := t.TaskSpec()
		rtr, err := resources.ResolveTaskResources(&spec, t.TaskMetadata().Name, inputs, outputs, getResource)
		if err != nil {
			return nil, &ResourceNotFoundError{Msg: err.Error()}
		}
		rprt.ResolvedTaskResources = rtr

		// Add this task to the state of the PipelineRun
		state = append(state, &rprt)
	}
	return state, nil
}

// ResolveTaskRuns will go through all tasks in state and check if there are existing TaskRuns
// for each of them by calling getTaskRun.
func ResolveTaskRuns(getTaskRun GetTaskRun, state []*ResolvedPipelineRunTask) error {
	for _, rprt := range state {
		// Check if we have already started a TaskRun for this task
		taskRun, err := getTaskRun(rprt.TaskRunName)
		if err != nil {
			// If the TaskRun isn't found, it just means it hasn't been run yet
			if !errors.IsNotFound(err) {
				return fmt.Errorf("error retrieving TaskRun %s: %s", rprt.TaskRunName, err)
			}
		} else {
			rprt.TaskRun = taskRun
		}
	}
	return nil
}

// getTaskRunName should return a uniquie name for a `TaskRun`.
func getTaskRunName(taskRunsStatus map[string]v1alpha1.TaskRunStatus, prName string, pt *v1alpha1.PipelineTask) string {
	base := fmt.Sprintf("%s-%s", prName, pt.Name)

	for k := range taskRunsStatus {
		if strings.HasPrefix(k, base) {
			return k
		}
	}

	return names.SimpleNameGenerator.GenerateName(base)
}

// GetPipelineConditionStatus will return the Condition that the PipelineRun prName should be
// updated with, based on the status of the TaskRuns in state.
func GetPipelineConditionStatus(prName string, state []*ResolvedPipelineRunTask, logger *zap.SugaredLogger, startTime *metav1.Time,
	pipelineTimeout *metav1.Duration) *duckv1alpha1.Condition {
	allFinished := true
	if !startTime.IsZero() && pipelineTimeout != nil {
		timeout := pipelineTimeout.Duration
		runtime := time.Since(startTime.Time)
		if runtime > timeout {
			logger.Infof("PipelineRun %q has timed out(runtime %s over %s)", prName, runtime, timeout)

			timeoutMsg := fmt.Sprintf("PipelineRun %q failed to finish within %q", prName, timeout.String())
			return &duckv1alpha1.Condition{
				Type:    duckv1alpha1.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  ReasonTimedOut,
				Message: timeoutMsg,
			}
		}
	}
	for _, rprt := range state {
		if rprt.TaskRun == nil {
			logger.Infof("TaskRun %s doesn't have a Status, so PipelineRun %s isn't finished", rprt.TaskRunName, prName)
			allFinished = false
			break
		}
		c := rprt.TaskRun.Status.GetCondition(duckv1alpha1.ConditionSucceeded)
		if c == nil {
			logger.Infof("TaskRun %s doesn't have a condition, so PipelineRun %s isn't finished", rprt.TaskRunName, prName)
			allFinished = false
			break
		}
		// If any TaskRuns have failed, we should halt execution and consider the run failed
		if c.Status == corev1.ConditionFalse {
			logger.Infof("TaskRun %s has failed, so PipelineRun %s has failed", rprt.TaskRunName, prName)
			return &duckv1alpha1.Condition{
				Type:    duckv1alpha1.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  ReasonFailed,
				Message: fmt.Sprintf("TaskRun %s has failed", rprt.TaskRun.Name),
			}
		}
		if c.Status != corev1.ConditionTrue {
			logger.Infof("TaskRun %s is still running so PipelineRun %s is still running", rprt.TaskRunName, prName)
			allFinished = false
		}
	}
	if !allFinished {
		logger.Infof("PipelineRun %s still has running TaskRuns so it isn't yet done", prName)
		return &duckv1alpha1.Condition{
			Type:    duckv1alpha1.ConditionSucceeded,
			Status:  corev1.ConditionUnknown,
			Reason:  ReasonRunning,
			Message: "Not all Tasks in the Pipeline have finished executing",
		}
	}
	logger.Infof("All TaskRuns have finished for PipelineRun %s so it has finished", prName)
	return &duckv1alpha1.Condition{
		Type:    duckv1alpha1.ConditionSucceeded,
		Status:  corev1.ConditionTrue,
		Reason:  ReasonSucceeded,
		Message: "All Tasks have completed executing",
	}
}

func findReferencedTask(pb string, state []*ResolvedPipelineRunTask) *ResolvedPipelineRunTask {
	for _, rprtRef := range state {
		if rprtRef.PipelineTask.Name == pb {
			return rprtRef
		}
	}
	return nil
}

// ValidateFrom will look at any `from` clauses in the resolved PipelineRun state
// and validate it: the `from` must specify an input of the current `Task`. The `PipelineTask`
// it corresponds to must actually exist in the `Pipeline`. The `PipelineResource` that is bound to the input
// must be the same `PipelineResource` that was bound to the output of the previous `Task`. If the state is
// not valid, it will return an error.
func ValidateFrom(state []*ResolvedPipelineRunTask) error {
	for _, rprt := range state {
		if rprt.PipelineTask.Resources != nil {
			for _, dep := range rprt.PipelineTask.Resources.Inputs {
				inputBinding := rprt.ResolvedTaskResources.Inputs[dep.Name]
				for _, pb := range dep.From {
					if pb == rprt.PipelineTask.Name {
						return fmt.Errorf("PipelineTask %s is trying to depend on a PipelineResource from itself", pb)
					}
					depTask := findReferencedTask(pb, state)
					if depTask == nil {
						return fmt.Errorf("pipelineTask %s is trying to depend on previous Task %q but it does not exist", rprt.PipelineTask.Name, pb)
					}

					sameBindingExists := false
					for _, output := range depTask.ResolvedTaskResources.Outputs {
						if output.Name == inputBinding.Name {
							sameBindingExists = true
						}
					}
					if !sameBindingExists {
						return fmt.Errorf("from is ambiguous: input %q for PipelineTask %q is bound to %q but no outputs in PipelineTask %q are bound to same resource",
							dep.Name, rprt.PipelineTask.Name, inputBinding.Name, depTask.PipelineTask.Name)
					}
				}
			}
		}
	}

	return nil
}
