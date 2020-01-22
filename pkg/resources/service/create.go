// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/ghodss/yaml"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// Deploy receives Service structure and generate knative/service object to deploy it in knative cluster
func (s *Service) Deploy(clientset *client.ConfigSet) (string, error) {
	var err error
	service := &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
	}

	image := s.Source
	builder := NewBuilder(clientset, s)

	if builder != nil && !client.Dry {
		defer func() {
			owner := metav1.OwnerReference{
				APIVersion: "serving.knative.dev/v1",
				Kind:       "Configuration",
				Name:       service.GetName(),
				UID:        service.GetUID(),
			}
			if err := builder.SetOwner(clientset, owner); err != nil {
				// fmt.Printf("Can't set builder owner\n")
				// if err = builder.Delete(clientset); err != nil {
				// fmt.Printf("Can't remove builder: %s\n", err)
				// }
			}
		}()

		if image, err = builder.Deploy(clientset); err != nil {
			return "", fmt.Errorf("Deploying builder: %s", err)
		}
	}
	clientset.Log.Debugf("image is ready, creating service\n")

	if s.BuildOnly {
		return fmt.Sprintf("Build-only flag set, service image is %s\n", image), nil
	}

	concurrency := int64(s.Concurrency)
	configuration := servingv1.ConfigurationSpec{
		Template: servingv1.RevisionTemplateSpec{
			Spec: servingv1.RevisionSpec{
				ContainerConcurrency: &concurrency,
				PodSpec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: image},
					},
				},
			},
		},
	}

	configuration.Template.ObjectMeta = metav1.ObjectMeta{
		CreationTimestamp: metav1.Time{Time: time.Now()},
		Annotations:       s.Annotations,
		Labels:            mapFromSlice(s.Labels),
	}

	configuration.Template.ObjectMeta.GenerateName = s.Name + "-"
	configuration.Template.ObjectMeta.Namespace = s.Namespace
	configuration.Template.Spec.PodSpec.Containers[0].Env = s.setupEnv()
	configuration.Template.Spec.PodSpec.Containers[0].EnvFrom = s.setupEnvSecrets()
	configuration.Template.Spec.PodSpec.Containers[0].ImagePullPolicy = corev1.PullPolicy(s.PullPolicy)

	service.ObjectMeta = metav1.ObjectMeta{
		Name:              s.Name,
		Namespace:         s.Namespace,
		Labels:            configuration.Template.ObjectMeta.Labels,
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}
	service.Spec = servingv1.ServiceSpec{
		ConfigurationSpec: configuration,
	}

	if client.Dry {
		var obj []byte
		if client.Output == "yaml" {
			obj, err = yaml.Marshal(service)
		} else {
			obj, err = json.MarshalIndent(service, "", " ")
		}
		return string(obj), err
	}

	if service, err = s.createOrUpdate(service, clientset); err != nil {
		return "", fmt.Errorf("Creating service: %s", err)
	}

	// TODO Add cronjob yaml into --dry output
	// if len(s.Cronjob.Schedule) != 0 {
	// 	if err := s.CreateCronjobSource(clientset); err != nil {
	// 		return "", fmt.Errorf("Creating cronjob source: %s", err)
	// 	}
	// }

	if !client.Wait {
		return fmt.Sprintf("Deployment started. Run \"tm -n %s describe service %s\" to see details", s.Namespace, s.Name), nil
	}

	clientset.Log.Infof("Waiting for service %q ready state\n", s.Name)
	domain, err := s.wait(clientset)
	return fmt.Sprintf("Service %s URL: %s", s.Name, domain), err
}

func (s *Service) setupEnv() []corev1.EnvVar {
	var env []corev1.EnvVar
	for k, v := range mapFromSlice(s.Env) {
		env = append(env, corev1.EnvVar{Name: k, Value: v})
	}
	return env
}

func (s *Service) setupEnvSecrets() []corev1.EnvFromSource {
	optional := true
	env := []corev1.EnvFromSource{}
	for _, secret := range s.EnvSecrets {
		env = append(env, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret,
				},
				Optional: &optional,
			},
		})
	}
	return env
}

func (s *Service) createOrUpdate(serviceObject *servingv1.Service, clientset *client.ConfigSet) (*servingv1.Service, error) {
	clientset.Log.Debugf("creating \"%s/%s\" service\n", s.Namespace, s.Name)
	newService, err := clientset.Serving.ServingV1().Services(s.Namespace).Create(serviceObject)
	if k8sErrors.IsAlreadyExists(err) {
		clientset.Log.Debugf("service \"%s/%s\" already exist, updating\n", serviceObject.GetNamespace(), serviceObject.GetName())
		service, err := clientset.Serving.ServingV1().Services(s.Namespace).Get(serviceObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if creator, exist := service.GetAnnotations()["serving.knative.dev/creator"]; exist {
			if serviceObject.Annotations == nil {
				ann := make(map[string]string)
				ann["serving.knative.dev/creator"] = creator
				serviceObject.SetAnnotations(ann)
			} else {
				serviceObject.Annotations["serving.knative.dev/creator"] = creator
			}
		}
		serviceObject.ObjectMeta.ResourceVersion = service.GetResourceVersion()
		return clientset.Serving.ServingV1().Services(s.Namespace).Update(serviceObject)
	}
	return newService, err
}

func mapFromSlice(slice []string) map[string]string {
	m := make(map[string]string)
	for _, s := range slice {
		t := regexp.MustCompile("[:=]").Split(s, 2)
		if len(t) != 2 {
			fmt.Printf("Can't parse argument slice %s\n", s)
			continue
		}
		m[t[0]] = t[1]
	}
	return m
}

func (s *Service) wait(clientset *client.ConfigSet) (string, error) {
	svcWatchInterface, err := clientset.Serving.ServingV1().Services(s.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
	})
	if err != nil {
		return "", err
	}
	if svcWatchInterface == nil {
		return "", errors.New("can't get watch interface, please check service status")
	}
	defer svcWatchInterface.Stop()

	duration, err := time.ParseDuration(s.BuildTimeout)
	if err != nil {
		duration = 10 * time.Minute
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	firstError := true
	for {
		select {
		case event := <-svcWatchInterface.ResultChan():
			if event.Object == nil {
				svcWatchInterface.Stop()
				if svcWatchInterface, err = clientset.Serving.ServingV1().Services(s.Namespace).Watch(metav1.ListOptions{
					FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
				}); err != nil {
					return "", err
				}
				if svcWatchInterface == nil {
					return "", errors.New("can't get watch interface, please check service status")
				}
				continue
			}
			serviceEvent, ok := event.Object.(*servingv1.Service)
			if !ok {
				continue
			}
			if clientset.Log.IsDebug() {
				clientset.Log.Debugf("got new event:\n")
				for _, v := range serviceEvent.Status.Conditions {
					clientset.Log.Debugf(" condition: %q, status: %q, message: %q\n", v.Type, v.Status, v.Message)
				}
			}
			if serviceEvent.Status.IsReady() {
				return serviceEvent.Status.URL.String(), nil
			}
			for _, v := range serviceEvent.Status.Conditions {
				if v.IsFalse() && v.Severity == apis.ConditionSeverityError {
					if v.Reason == "RevisionFailed" && firstError {
						time.Sleep(time.Second * 3)
						svcWatchInterface.Stop()
						if svcWatchInterface, err = clientset.Serving.ServingV1().Services(s.Namespace).Watch(metav1.ListOptions{
							FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
						}); err != nil {
							return "", err
						}
						if svcWatchInterface == nil {
							return "", errors.New("can't get watch interface, please check service status")
						}
						firstError = false
						break
					}
					return "", errors.New(v.Message)
				}
			}
		case <-ticker.C:
			return "", fmt.Errorf("watch service timeout")
		}
	}
}
