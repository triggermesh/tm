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
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/resources/build"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy receives Service structure and generate knative/service object to deploy it in knative cluster
func (s *Service) Deploy(clientset *client.ConfigSet) (string, error) {
	fmt.Printf("Creating %s function\n", s.Name)
	var configuration servingv1alpha1.ConfigurationSpec

	build := build.Build{
		Name:           s.Name,
		Source:         s.Source,
		Registry:       s.Registry,
		Args:           s.BuildArgs,
		Namespace:      s.Namespace,
		Timeout:        s.BuildTimeout,
		Buildtemplate:  s.Buildtemplate,
		RegistrySecret: s.RegistrySecret,
	}

	image, err := build.Deploy(clientset)
	if err != nil {
		return "", err
	}
	// if s.Buildtemplate != "" {
	// 	if newBuildtemplate, err = s.cloneBuildtemplate(clientset); err != nil {
	// 		return "", fmt.Errorf("Creating temporary buildtemplate: %s", err)
	// 	} else if newBuildtemplate == nil {
	// 		rand, err := uniqueString()
	// 		if err != nil {
	// 			return "", fmt.Errorf("Generating unique buildtemplate name: %s", err)
	// 		}
	// 		b := buildtemplate.Buildtemplate{
	// 			Name:           fmt.Sprintf("%s-%s", s.Name, rand),
	// 			Namespace:      s.Namespace,
	// 			File:           s.Buildtemplate,
	// 			RegistrySecret: s.RegistrySecret,
	// 		}
	// 		if newBuildtemplate, err = b.Deploy(clientset); err != nil {
	// 			return "", fmt.Errorf("Deploying new buildtemplate: %s", err)
	// 		}
	// 	}
	// 	s.Buildtemplate = newBuildtemplate.GetName()
	// }

	// switch {
	// case file.IsLocal(s.Source):
	// 	if file.IsDir(s.Source) {
	// 		s.Source = path.Clean(s.Source)
	// 	} else {
	// 		s.BuildArgs = append(s.BuildArgs, "HANDLER="+path.Base(s.Source))
	// 		s.Source = path.Clean(path.Dir(s.Source))
	// 	}
	// 	s.BuildArgs = append(s.BuildArgs, "DIRECTORY=.")
	// 	build.Spec = s.buildPath()
	// case file.IsGit(s.Source):
	// 	if len(s.Revision) == 0 {
	// 		s.Revision = "master"
	// 	}
	// 	build.Spec = s.buildSource()
	// default:
	// 	configuration = s.configFromImage()
	// }

	// image, err := s.imageName(clientset)
	// if err != nil {
	// 	return "", fmt.Errorf("Composing service image name: %s", err)
	// }

	// timeout, err := time.ParseDuration(s.BuildTimeout)
	// if err != nil {
	// 	timeout = 10 * time.Minute
	// }

	// if build.Spec.Source != nil {
	// build.Spec.Timeout = &metav1.Duration{Duration: timeout}
	// build.Spec.Template = &buildv1alpha1.TemplateInstantiationSpec{
	// 	Name:      s.Buildtemplate,
	// 	Arguments: getBuildArguments(image, s.BuildArgs),
	// 	Env: []corev1.EnvVar{
	// 		{Name: "timestamp", Value: time.Now().String()},
	// 	},
	// }
	// if build, err = clientset.Build.BuildV1alpha1().Builds(s.Namespace).Create(build); err != nil {
	// 	return "", fmt.Errorf("Service build error: %s", err)
	// }
	configuration.Template = &servingv1alpha1.RevisionTemplateSpec{
		Spec: servingv1alpha1.RevisionSpec{
			RevisionSpec: servingv1beta1.RevisionSpec{
				PodSpec: servingv1beta1.PodSpec{
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
	configuration.Template.Spec.ContainerConcurrency = servingv1beta1.RevisionContainerConcurrencyType(s.Concurrency)
	configuration.Template.Spec.RevisionSpec.PodSpec.Containers[0].Env = s.setupEnv()
	configuration.Template.Spec.RevisionSpec.PodSpec.Containers[0].EnvFrom = s.setupEnvSecrets()
	configuration.Template.Spec.RevisionSpec.PodSpec.Containers[0].ImagePullPolicy = corev1.PullPolicy(s.PullPolicy)

	serviceObject := servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:              s.Name,
			Namespace:         s.Namespace,
			Labels:            configuration.Template.ObjectMeta.Labels,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: servingv1alpha1.ServiceSpec{
			ConfigurationSpec: configuration,
		},
	}

	if client.Dry {
		var obj []byte
		if client.Output == "yaml" {
			obj, err = yaml.Marshal(serviceObject)
		} else {
			obj, err = json.MarshalIndent(serviceObject, "", " ")
		}
		return string(obj), err
	}

	if _, err := s.createOrUpdate(serviceObject, clientset); err != nil {
		return "", fmt.Errorf("Creating service: %s", err)
	}

	// if build.Spec.Source != nil {
	// 	if build, err = clientset.Build.BuildV1alpha1().Builds(s.Namespace).Get(build.Name, metav1.GetOptions{}); err != nil {
	// 		return "", err
	// 	}
	// 	conf, err := clientset.Serving.ServingV1alpha1().Configurations(s.Namespace).Get(s.Name, metav1.GetOptions{})
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	trueP := true
	// 	ref := metav1.OwnerReference{
	// 		APIVersion:         "serving.knative.dev/v1alpha1",
	// 		Kind:               "Configuration",
	// 		Name:               conf.GetName(),
	// 		UID:                conf.GetUID(),
	// 		BlockOwnerDeletion: &trueP,
	// 		Controller:         &trueP,
	// 	}
	// 	build.SetOwnerReferences([]metav1.OwnerReference{ref})
	// 	if _, err := clientset.Build.BuildV1alpha1().Builds(s.Namespace).Update(build); err != nil {
	// 		return "", err
	// 	}
	// }

	// if s.Buildtemplate != "" {
	// 	if err := s.setBuildtemplateOwner(newBuildtemplate, newService, clientset); err != nil {
	// 		return "", fmt.Errorf("Setting buildtemplate owner: %s", err)
	// 	}
	// }

	// if file.IsLocal(s.Source) {
	// 	if err := s.injectSources(clientset); err != nil {
	// 		return "", fmt.Errorf("Injecting service sources: %s", err)
	// 	}
	// }

	// TODO Add cronjob yaml into --dry output
	if len(s.Cronjob.Schedule) != 0 {
		if err := s.CreateCronjobSource(clientset); err != nil {
			return "", fmt.Errorf("Creating cronjob source: %s", err)
		}
	}

	if !client.Wait {
		return fmt.Sprintf("Deployment started. Run \"tm -n %s describe service %s\" to see details", s.Namespace, s.Name), nil
	}

	fmt.Printf("Waiting for %s ready state\n", s.Name)
	domain, err := s.waitService(clientset)
	if err != nil {
		return "", fmt.Errorf("Waiting for service readiness: %s", err)
	}
	return fmt.Sprintf("Service %s URL: http://%s", s.Name, domain), nil
}

func (s *Service) setupEnv() []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  "timestamp",
			Value: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
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

func (s *Service) createOrUpdate(serviceObject servingv1alpha1.Service, clientset *client.ConfigSet) (*servingv1alpha1.Service, error) {
	newService, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).Create(&serviceObject)
	if k8sErrors.IsAlreadyExists(err) {
		service, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).Get(serviceObject.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		serviceObject.ObjectMeta.ResourceVersion = service.GetResourceVersion()
		return clientset.Serving.ServingV1alpha1().Services(s.Namespace).Update(&serviceObject)
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

func (s *Service) waitService(clientset *client.ConfigSet) (string, error) {
	res, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
	})
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", errors.New("can't get watch interface, please check service status")
	}
	defer res.Stop()

	firstError := true
	for {
		event := <-res.ResultChan()
		if event.Object == nil {
			res.Stop()
			if res, err = clientset.Serving.ServingV1alpha1().Services(s.Namespace).Watch(metav1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
			}); err != nil {
				return "", err
			}
			if res == nil {
				return "", errors.New("can't get watch interface, please check service status")
			}
			continue
		}
		serviceEvent, ok := event.Object.(*servingv1alpha1.Service)
		if !ok {
			continue
		}
		if serviceEvent.Status.IsReady() {
			return serviceEvent.Status.Domain, nil
		}
		for _, v := range serviceEvent.Status.Conditions {
			if v.IsFalse() {
				if v.Reason == "RevisionFailed" && firstError {
					time.Sleep(time.Second * 3)
					res.Stop()
					if res, err = clientset.Serving.ServingV1alpha1().Services(s.Namespace).Watch(metav1.ListOptions{
						FieldSelector: fmt.Sprintf("metadata.name=%s", s.Name),
					}); err != nil {
						return "", err
					}
					if res == nil {
						return "", errors.New("can't get watch interface, please check service status")
					}
					firstError = false
					break
				}
				return "", errors.New(v.Message)
			}
		}
	}
}

// func addSecretVolume(registrySecret string, template *buildv1alpha1.BuildTemplate) {
// 	template.Spec.Volumes = []corev1.Volume{
// 		{
// 			Name: registrySecret,
// 			VolumeSource: corev1.VolumeSource{
// 				Secret: &corev1.SecretVolumeSource{
// 					SecretName: registrySecret,
// 				},
// 			},
// 		},
// 	}
// 	for i, step := range template.Spec.Steps {
// 		mounts := append(step.VolumeMounts, corev1.VolumeMount{
// 			Name:      registrySecret,
// 			MountPath: "/" + registrySecret,
// 			ReadOnly:  true,
// 		})
// 		template.Spec.Steps[i].VolumeMounts = mounts
// 	}
// }

// func setEnvConfig(registrySecret string, template *buildv1alpha1.BuildTemplate) {
// 	for i, step := range template.Spec.Steps {
// 		envs := append(step.Env, corev1.EnvVar{
// 			Name:  "DOCKER_CONFIG",
// 			Value: "/" + registrySecret,
// 		})
// 		template.Spec.Steps[i].Env = envs
// 	}
// }
