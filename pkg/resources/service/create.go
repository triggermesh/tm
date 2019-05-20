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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/buildtemplate"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const uploadDoneTrigger = "/home/.sourceuploaddone"

// Deploy receives Service structure and generate knative/service object to deploy it in knative cluster
func (s *Service) Deploy(clientset *client.ConfigSet) (string, error) {
	fmt.Printf("Creating %s function\n", s.Name)
	var configuration servingv1alpha1.ConfigurationSpec
	var newBuildtemplate *buildv1alpha1.BuildTemplate
	var err error

	if s.Buildtemplate != "" {
		if newBuildtemplate, err = s.cloneBuildtemplate(clientset); err != nil {
			return "", fmt.Errorf("Creating temporary buildtemplate: %s", err)
		} else if newBuildtemplate == nil {
			rand, err := uniqueString()
			if err != nil {
				return "", fmt.Errorf("Generating unique buildtemplate name: %s", err)
			}
			b := buildtemplate.Buildtemplate{
				Name:           fmt.Sprintf("%s-%s", s.Name, rand),
				Namespace:      s.Namespace,
				File:           s.Buildtemplate,
				RegistrySecret: s.RegistrySecret,
			}
			if newBuildtemplate, err = b.Deploy(clientset); err != nil {
				return "", fmt.Errorf("Deploying new buildtemplate: %s", err)
			}
		}
		s.Buildtemplate = newBuildtemplate.GetName()
	}

	switch {
	case file.IsLocal(s.Source):
		if file.IsDir(s.Source) {
			s.Source = path.Clean(s.Source)
		} else {
			s.BuildArgs = append(s.BuildArgs, "HANDLER="+path.Base(s.Source))
			s.Source = path.Clean(path.Dir(s.Source))
		}
		s.BuildArgs = append(s.BuildArgs, "DIRECTORY=.")
		configuration = s.fromPath()
	case file.IsGit(s.Source):
		if len(s.Revision) == 0 {
			s.Revision = "master"
		}
		configuration = s.fromSource()
	default:
		configuration = s.fromImage()
	}

	image, err := s.imageName(clientset)
	if err != nil {
		return "", fmt.Errorf("Composing service image name: %s", err)
	}

	timeout, err := time.ParseDuration(s.BuildTimeout)
	if err != nil {
		timeout = 10 * time.Minute
	}

	if configuration.Build != nil {
		configuration.RevisionTemplate = servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: fmt.Sprintf("%s:%s", image, s.ResultImageTag),
				},
			},
		}
		if configuration.Build.BuildSpec != nil {
			configuration.Build.BuildSpec.Timeout = &metav1.Duration{timeout}
			configuration.Build.BuildSpec.Template = &buildv1alpha1.TemplateInstantiationSpec{
				Name:      s.Buildtemplate,
				Arguments: getBuildArguments(image, s.BuildArgs),
				Env: []corev1.EnvVar{
					{Name: "timestamp", Value: time.Now().String()},
				},
			}
		}
	}

	configuration.RevisionTemplate.ObjectMeta = metav1.ObjectMeta{
		CreationTimestamp: metav1.Time{time.Now()},
		Annotations:       s.Annotations,
		Labels:            mapFromSlice(s.Labels),
	}

	configuration.RevisionTemplate.Spec.ContainerConcurrency = servingv1alpha1.RevisionContainerConcurrencyType(s.Concurrency)
	configuration.RevisionTemplate.Spec.Container.Env = s.setupEnv()
	configuration.RevisionTemplate.Spec.Container.EnvFrom = s.setupEnvSecrets()
	configuration.RevisionTemplate.Spec.Container.ImagePullPolicy = corev1.PullPolicy(s.PullPolicy)

	serviceObject := servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:              s.Name,
			Namespace:         s.Namespace,
			Labels:            configuration.RevisionTemplate.ObjectMeta.Labels,
			CreationTimestamp: metav1.Time{time.Now()},
		},
		Spec: servingv1alpha1.ServiceSpec{
			RunLatest: &servingv1alpha1.RunLatestType{
				Configuration: configuration,
			},
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

	newService, err := s.createOrUpdate(serviceObject, clientset)
	if err != nil {
		return "", fmt.Errorf("Creating service: %s", err)
	}

	if s.Buildtemplate != "" {
		if err := s.setBuildtemplateOwner(newBuildtemplate, newService, clientset); err != nil {
			return "", fmt.Errorf("Setting buildtemplate owner: %s", err)
		}
	}

	if file.IsLocal(s.Source) {
		if err := s.injectSources(clientset); err != nil {
			return "", fmt.Errorf("Injecting service sources: %s", err)
		}
	}

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
	return fmt.Sprintf("Service %s URL: http://%s\n", s.Name, domain), nil
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

func (s *Service) cloneBuildtemplate(clientset *client.ConfigSet) (*buildv1alpha1.BuildTemplate, error) {
	rand, err := uniqueString()
	if err != nil {
		return nil, err
	}
	bt := buildtemplate.Buildtemplate{
		Name:           fmt.Sprintf("%s-%s", s.Name, rand),
		Namespace:      s.Namespace,
		RegistrySecret: s.RegistrySecret,
	}

	sourceBt, err := clientset.Build.BuildV1alpha1().BuildTemplates(s.Namespace).Get(s.Buildtemplate, metav1.GetOptions{})
	if err != nil {
		cb, err := clientset.Build.BuildV1alpha1().ClusterBuildTemplates().Get(s.Buildtemplate, metav1.GetOptions{})
		if err != nil {
			return nil, nil
		}
		sourceBt.Spec = cb.Spec
		sourceBt.TypeMeta = cb.TypeMeta
		sourceBt.ObjectMeta = cb.ObjectMeta
	}
	return bt.Clone(*sourceBt, clientset)
}

func (s *Service) setBuildtemplateOwner(buildtemplate *buildv1alpha1.BuildTemplate, owner *servingv1alpha1.Service, clientset *client.ConfigSet) error {
	buildtemplate.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
			Name:       s.Name,
			UID:        owner.GetUID(),
		},
	})
	_, err := clientset.Build.BuildV1alpha1().BuildTemplates(s.Namespace).Update(buildtemplate)
	return err
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

func (s *Service) fromImage() servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: s.Source,
				},
			},
		},
	}
}

func (s *Service) fromSource() servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		Build: &servingv1alpha1.RawExtension{
			BuildSpec: &buildv1alpha1.BuildSpec{
				Source: &buildv1alpha1.SourceSpec{
					Git: &buildv1alpha1.GitSourceSpec{
						Url:      s.Source,
						Revision: s.Revision,
					},
				},
			},
		},
	}
}

func (s *Service) fromPath() servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		Build: &servingv1alpha1.RawExtension{
			BuildSpec: &buildv1alpha1.BuildSpec{
				Source: &buildv1alpha1.SourceSpec{
					Custom: &corev1.Container{
						Image:   "library/busybox",
						Command: []string{"sh"},
						Args: []string{"-c", fmt.Sprintf(`
						while [ ! -f %s ]; do 
							sleep 1; 
						done; 
						sync; 
						mv /home/%s/* /workspace; 
						sync;`,
							uploadDoneTrigger, path.Base(s.Source))},
					},
				},
			},
		},
	}
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

func (s *Service) inProgress(clientset *client.ConfigSet) bool {
	service, err := clientset.Serving.ServingV1alpha1().Services(s.Namespace).Get(s.Name, metav1.GetOptions{})
	if err != nil {
		return false
	}
	cond := service.Status.GetCondition(servingv1alpha1.ServiceConditionReady)
	return cond != nil && cond.Status == corev1.ConditionUnknown
}

func (s *Service) buildPodName(clientset *client.ConfigSet) (string, error) {
	list, err := clientset.Build.BuildV1alpha1().Builds(s.Namespace).List(metav1.ListOptions{
		IncludeUninitialized: true,
	})
	if err != nil {
		return "", err
	}
	var builds []buildv1alpha1.Build
	for _, build := range list.Items {
		if len(build.OwnerReferences) == 1 &&
			build.OwnerReferences[0].Kind == "Configuration" &&
			build.OwnerReferences[0].Name == s.Name {
			builds = append(builds, build)
		}
	}
	var latest string
	var timestamp time.Time
	for _, build := range builds {
		cond := build.Status.GetCondition(buildv1alpha1.BuildSucceeded)
		if cond != nil && cond.Status == corev1.ConditionUnknown {
			if build.Status.StartTime != nil && build.Status.StartTime.After(timestamp) {
				if build.Status.Cluster != nil {
					timestamp = build.Status.StartTime.Time
					latest = build.Status.Cluster.PodName
				}
			}
		}
	}
	return latest, nil
}

func (s *Service) injectSources(clientset *client.ConfigSet) error {
	var buildPod string
	var err error
	for buildPod == "" {
		if buildPod, err = s.buildPodName(clientset); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("Uploading sources to %s\n", buildPod)
	res, err := clientset.Core.CoreV1().Pods(s.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod})
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("can't get watch interface, please check build status")
	}
	defer res.Stop()

	var sourceContainer string
	for sourceContainer == "" {
		event := <-res.ResultChan()
		if event.Object == nil {
			res.Stop()
			if res, err = clientset.Core.CoreV1().Pods(s.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod}); err != nil {
				return err
			}
			if res == nil {
				return errors.New("can't get watch interface, please check build status")
			}
			continue
		}
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		for _, v := range pod.Status.InitContainerStatuses {
			if v.Name == "build-step-custom-source" {
				if v.State.Terminated != nil {
					// Looks like we got watch interface for "previous" service version
					// Trying to get latest build pod name
					for buildPod = ""; buildPod == "" && s.inProgress(clientset); {
						if buildPod, err = s.buildPodName(clientset); err != nil {
							return err
						}
						time.Sleep(2 * time.Second)
					}
					if buildPod == "" {
						return fmt.Errorf("Can't get build pod name, please check service status")
					}
					res.Stop()
					fmt.Printf("Updating build pod name to %s\n", buildPod)
					if res, err = clientset.Core.CoreV1().Pods(s.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod}); err != nil {
						return err
					}
					break
				}
				if v.State.Running != nil {
					sourceContainer = v.Name
					break
				}
			}
		}
	}

	c := file.Copy{
		Pod:         buildPod,
		Namespace:   s.Namespace,
		Container:   sourceContainer,
		Source:      s.Source,
		Destination: path.Join("/home", path.Base(s.Source)),
	}
	if err := c.Upload(clientset); err != nil {
		return err
	}

	if _, _, err := c.RemoteExec(clientset, "touch "+uploadDoneTrigger, nil); err != nil {
		return err
	}

	return nil
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

func (s *Service) imageName(clientset *client.ConfigSet) (string, error) {
	if len(s.RegistrySecret) == 0 {
		return fmt.Sprintf("%s/%s/%s", s.Registry, s.Namespace, s.Name), nil
	}
	secret, err := clientset.Core.CoreV1().Secrets(s.Namespace).Get(s.RegistrySecret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	data := secret.Data["config.json"]
	dec := json.NewDecoder(strings.NewReader(string(data)))
	var config registryAuths
	if err := dec.Decode(&config); err != nil {
		return "", err
	}
	if len(config.Auths) > 1 {
		return "", errors.New("credentials with multiple registries not supported")
	}
	for k, v := range config.Auths {
		if url, ok := gitlabEnv(); ok {
			return fmt.Sprintf("%s/%s", url, s.Name), nil
		}
		return fmt.Sprintf("%s/%s/%s", k, v.Username, s.Name), nil
	}
	return "", errors.New("empty registry credentials")
}

// hack to use correct username in image URL instead of "gitlab-ci-token" in Gitlab CI
func gitlabEnv() (string, bool) {
	return os.LookupEnv("CI_REGISTRY_IMAGE")
}

func uniqueString() (string, error) {
	b, err := ioutil.ReadAll(io.LimitReader(rand.Reader, 3))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getBuildArguments(image string, buildArgs []string) []buildv1alpha1.ArgumentSpec {
	args := []buildv1alpha1.ArgumentSpec{
		{
			Name:  "IMAGE",
			Value: image,
		},
	}

	for k, v := range mapFromSlice(buildArgs) {
		args = append(args, buildv1alpha1.ArgumentSpec{
			Name: k, Value: v,
		})
	}

	return args
}

func addSecretVolume(registrySecret string, template *buildv1alpha1.BuildTemplate) {
	template.Spec.Volumes = []corev1.Volume{
		{
			Name: registrySecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registrySecret,
				},
			},
		},
	}
	for i, step := range template.Spec.Steps {
		mounts := append(step.VolumeMounts, corev1.VolumeMount{
			Name:      registrySecret,
			MountPath: "/" + registrySecret,
			ReadOnly:  true,
		})
		template.Spec.Steps[i].VolumeMounts = mounts
	}
}

func setEnvConfig(registrySecret string, template *buildv1alpha1.BuildTemplate) {
	for i, step := range template.Spec.Steps {
		envs := append(step.Env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/" + registrySecret,
		})
		template.Spec.Steps[i].Env = envs
	}
}
