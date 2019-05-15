// // Copyright 2019 TriggerMesh, Inc
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.

package service

// import (
// 	"fmt"
// 	"regexp"
// 	"strings"

// 	eventingv1alpha1 "github.com/knative/eventing-sources/pkg/apis/sources/v1alpha1"
// 	"github.com/triggermesh/tm/pkg/client"
// 	corev1 "k8s.io/api/core/v1"
// 	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func (s *Service) CreateCronjobSource(clientset *client.ConfigSet) error {
// 	crontab, err := parseRate(s.Cronjob.Schedule)
// 	if err != nil {
// 		return err
// 	}
// 	cronjob := eventingv1alpha1.CronJobSource{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      s.Name + "-cronjob",
// 			Namespace: s.Namespace,
// 		},
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "CronJobSource",
// 			APIVersion: "sources.eventing.knative.dev/v1alpha1",
// 		},
// 		Spec: eventingv1alpha1.CronJobSourceSpec{
// 			Schedule: crontab,
// 			Data:     s.Cronjob.Data,
// 			Sink: &corev1.ObjectReference{
// 				APIVersion: "serving.knative.dev/v1alpha1",
// 				Kind:       "Service",
// 				Name:       s.Name,
// 				Namespace:  s.Namespace,
// 			},
// 		},
// 	}
// 	if client.Dry {
// 		fmt.Println(cronjob)
// 		return nil
// 	}
// 	_, err = clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Create(&cronjob)
// 	if k8sErrors.IsAlreadyExists(err) {
// 		oldSource, err := clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Get(cronjob.Name, metav1.GetOptions{})
// 		if err != nil {
// 			return err
// 		}
// 		cronjob.ResourceVersion = oldSource.GetResourceVersion()
// 		if _, err = clientset.EventSources.SourcesV1alpha1().CronJobSources(s.Namespace).Update(&cronjob); err != nil {
// 			return err
// 		}
// 	} else if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func parseRate(schedule string) (string, error) {
// 	if !strings.HasPrefix(schedule, "rate") {
// 		return schedule, nil
// 	}
// 	rx := regexp.MustCompile(`\((.*?)\)`)
// 	rate := rx.FindStringSubmatch(schedule)
// 	if len(rate) != 2 {
// 		return "", fmt.Errorf("invalid rate format")
// 	}
// 	value := strings.Fields(rate[1])
// 	if len(value) != 2 {
// 		return "", fmt.Errorf("rate must contain two parameters")
// 	}
// 	switch string(value[1][0]) {
// 	case "m":
// 		return fmt.Sprintf("*/%s * * * *", value[0]), nil
// 	case "h":
// 		return fmt.Sprintf("* */%s * * *", value[0]), nil
// 	case "d":
// 		return fmt.Sprintf("* * */%s * *", value[0]), nil
// 	}
// 	return "", fmt.Errorf("unknown value \"%s\"", value[1])
// }
