package deploy

import (
	"fmt"

	eventingv1alpha1 "github.com/knative/eventing-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Service) CreateCronjobSource(clientset *client.ConfigSet) error {
	cronjob := eventingv1alpha1.CronJobSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name + "-cronjob",
			Namespace: s.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJobSource",
			APIVersion: "sources.eventing.knative.dev/v1alpha1",
		},
		Spec: eventingv1alpha1.CronJobSourceSpec{
			Schedule: s.Cronjob.Schedule,
			Data:     s.Cronjob.Data,
			Sink: &corev1.ObjectReference{
				APIVersion: "serving.knative.dev/v1alpha1",
				Kind:       "Service",
				Name:       s.Name,
				Namespace:  s.Namespace,
			},
		},
	}
	if client.Dry {
		fmt.Println(cronjob)
		return nil
	}
	_, err := clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Create(&cronjob)
	if k8sErrors.IsAlreadyExists(err) {
		oldSource, err := clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Get(cronjob.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		cronjob.ResourceVersion = oldSource.GetResourceVersion()
		if _, err = clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Update(&cronjob); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
