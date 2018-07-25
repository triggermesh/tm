package knative

import (
	knativeApi "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/build/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetBuildCustomResource will delete custom build object
func GetBuildCustomResource(buildClient versioned.Interface, buildName, ns string) (*knativeApi.Build, error) {
	return buildClient.BuildV1alpha1().Builds(ns).Get(buildName, metav1.GetOptions{})
}

// GetBuildListByNamespace return BuildList object for specified namespace
func GetBuildListByNamespace(buildClient versioned.Interface, ns string) (*knativeApi.BuildList, error) {
	return buildClient.BuildV1alpha1().Builds(ns).List(metav1.ListOptions{})
}
