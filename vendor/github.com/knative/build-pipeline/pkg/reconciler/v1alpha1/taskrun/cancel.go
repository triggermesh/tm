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

package taskrun

import (
	"fmt"

	"github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type logger interface {
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
}

// isCancelled returns true if the TaskRun's' spec indicates it is cancelled
func isCancelled(spec v1alpha1.TaskRunSpec) bool {
	return spec.Status == v1alpha1.TaskRunSpecStatusCancelled
}

// cancelTaskRun marks the TaskRun as cancelled and delete pods linked to it.
func cancelTaskRun(tr *v1alpha1.TaskRun, clientSet kubernetes.Interface, logger logger) error {
	logger.Warn("task run %q has been cancelled", tr.Name)
	tr.Status.SetCondition(&duckv1alpha1.Condition{
		Type:    duckv1alpha1.ConditionSucceeded,
		Status:  corev1.ConditionFalse,
		Reason:  "TaskRunCancelled",
		Message: fmt.Sprintf("TaskRun %q was cancelled", tr.Name),
	})

	if tr.Status.PodName == "" {
		logger.Warnf("task run %q has no pod running yet", tr.Name)
		return nil
	}

	if err := clientSet.CoreV1().Pods(tr.Namespace).Delete(tr.Status.PodName, &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
