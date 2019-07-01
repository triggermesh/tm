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

package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	"github.com/triggermesh/tm/pkg/file"
	"github.com/triggermesh/tm/pkg/resources/buildtemplate"
	"github.com/triggermesh/tm/pkg/resources/clusterbuildtemplate"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const uploadDoneTrigger = "/home/.sourceuploaddone"

// Deploy uses Build structure to generate and deploy knative build
func (b *Build) Deploy(clientset *client.ConfigSet) (string, error) {
	build := &buildv1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Build",
			APIVersion: "build.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:         b.Name,
			GenerateName: b.GenerateName,
			Namespace:    b.Namespace,
		},
	}

	// workaround, find a better solution
	if b.Name == "" {
		b.Name = strings.TrimRight(b.GenerateName, "-")
	}

	if b.Buildtemplate != "" && !client.Dry {
		buildtemplateName, err := b.installBuildTemplate(clientset)
		if err != nil {
			return "", fmt.Errorf("buildtemplate installation: %s", err)
		}
		b.Buildtemplate = buildtemplateName
		fmt.Printf("Buildtemplate %q installed\n", b.Buildtemplate)

		defer func() {
			owner := metav1.OwnerReference{
				APIVersion: "build.knative.dev/v1alpha1",
				Kind:       "Build",
				Name:       build.Name,
				UID:        build.UID,
			}
			if err := b.setBuildtemplateOwner(clientset, owner); err != nil {
				fmt.Printf("Can't set buildtemplate owner, cleaning up\n")
				bt := buildtemplate.Buildtemplate{
					Name:      b.Buildtemplate,
					Namespace: b.Namespace,
				}
				if err = bt.Delete(clientset); err != nil {
					fmt.Printf("Can't remove buildtemplate %q: %s\n", build.Name, err)
				}
			}
		}()
	}

	var local bool
	if file.IsLocal(b.Source) {
		local = true
		if file.IsDir(b.Source) {
			b.Source = path.Clean(b.Source)
		} else {
			b.Args = append(b.Args, "HANDLER="+path.Base(b.Source))
			b.Source = path.Clean(path.Dir(b.Source))
		}
		b.Args = append(b.Args, "DIRECTORY=.")
		build.Spec = b.buildSpecLocalPath()
	} else if file.IsGit(b.Source) {
		if len(b.Revision) == 0 {
			b.Revision = "master"
		}
		build.Spec = b.buildSpecSourceURL()
	} else {
		build.Spec = buildv1alpha1.BuildSpec{
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Arguments: []buildv1alpha1.ArgumentSpec{
					{Name: "IMAGE", Value: b.Source},
				},
			},
		}
		return "", nil
	}

	duration, err := time.ParseDuration(b.Timeout)
	if err != nil {
		duration = 10 * time.Minute
	}

	image, err := b.imageName(clientset)
	if err != nil {
		return "", fmt.Errorf("Composing service image name: %s", err)
	}
	tag := file.RandString(6)

	build.Spec.Timeout = &metav1.Duration{Duration: duration}
	build.Spec.Template = &buildv1alpha1.TemplateInstantiationSpec{
		Name:      b.Buildtemplate,
		Arguments: getBuildArguments(image, tag, b.Args),
	}

	if client.Dry {
		res, err := json.MarshalIndent(build, "", " ")
		return string(res), err
	}

	if build, err = clientset.Build.BuildV1alpha1().Builds(b.Namespace).Create(build); err != nil {
		return "", fmt.Errorf("Build %q creation error: %s", build.Name, err)
	}
	b.Name = build.GetName()

	if local {
		if err := b.injectSources(clientset); err != nil {
			return "", fmt.Errorf("Injecting sources: %s", err)
		}
	}

	if b.Wait {
		fmt.Printf("Waiting for the build %q completion\n", b.Name)
		if err := b.wait(build, clientset); err != nil {
			return "", fmt.Errorf("Build %q error: %s", b.Name, err)
		}
	}

	return fmt.Sprintf("%s:%s", image, tag), nil
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

func (b *Build) injectSources(clientset *client.ConfigSet) error {
	var buildPod string
	var err error
	for buildPod == "" {
		if buildPod, err = b.buildPodName(clientset); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("Uploading sources to %q\n", buildPod)
	res, err := clientset.Core.CoreV1().Pods(b.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod})
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
			if res, err = clientset.Core.CoreV1().Pods(b.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod}); err != nil {
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
					for buildPod = ""; buildPod == "" && b.inProgress(clientset); {
						if buildPod, err = b.buildPodName(clientset); err != nil {
							return err
						}
						time.Sleep(2 * time.Second)
					}
					if buildPod == "" {
						return fmt.Errorf("Can't get build pod name, please check service status")
					}
					res.Stop()
					fmt.Printf("Updating build pod name to %s\n", buildPod)
					if res, err = clientset.Core.CoreV1().Pods(b.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + buildPod}); err != nil {
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
		Namespace:   b.Namespace,
		Container:   sourceContainer,
		Source:      b.Source,
		Destination: path.Join("/home", path.Base(b.Source)),
	}
	if err := c.Upload(clientset); err != nil {
		return err
	}

	if _, _, err := c.RemoteExec(clientset, "touch "+uploadDoneTrigger, nil); err != nil {
		return err
	}

	return nil
}

func (b *Build) buildPodName(clientset *client.ConfigSet) (string, error) {
	list, err := clientset.Build.BuildV1alpha1().Builds(b.Namespace).List(metav1.ListOptions{
		FieldSelector: "metadata.name=" + b.Name,
	})
	if err != nil {
		return "", err
	}
	var latestBuild string
	var latestStart time.Time
	for _, build := range list.Items {
		if b.inProgress(clientset) {
			if build.Status.StartTime != nil && build.Status.StartTime.After(latestStart) {
				if build.Status.Cluster != nil {
					latestStart = build.Status.StartTime.Time
					latestBuild = build.Status.Cluster.PodName
				}
			}
		}
	}
	return latestBuild, nil
}

func (b *Build) inProgress(clientset *client.ConfigSet) bool {
	build, err := clientset.Build.BuildV1alpha1().Builds(b.Namespace).Get(b.Name, metav1.GetOptions{})
	if err != nil {
		return false
	}
	cond := build.Status.GetCondition(buildv1alpha1.BuildSucceeded)
	return cond != nil && cond.Status == corev1.ConditionUnknown
}

func (b *Build) installBuildTemplate(clientset *client.ConfigSet) (string, error) {
	newBuildtemplate, err := b.cloneBuildtemplate(clientset)
	if err != nil {
		return "", err
	} else if newBuildtemplate == nil {
		bt := buildtemplate.Buildtemplate{
			File:           b.Buildtemplate,
			Namespace:      b.Namespace,
			RegistrySecret: b.RegistrySecret,
		}
		if newBuildtemplate, err = bt.Deploy(clientset); err != nil {
			return "", err
		}
		b.Buildtemplate = newBuildtemplate.Name
		newBuildtemplate, err = b.cloneBuildtemplate(clientset)
		if err != nil {
			return "", err
		}
	}
	return newBuildtemplate.GetName(), nil
}

func (b *Build) setBuildtemplateOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	bt := buildtemplate.Buildtemplate{
		Name:      b.Buildtemplate,
		Namespace: b.Namespace,
	}
	return bt.SetOwner(clientset, owner)
}

func (b *Build) cloneBuildtemplate(clientset *client.ConfigSet) (*buildv1alpha1.BuildTemplate, error) {
	bt := buildtemplate.Buildtemplate{
		Name:           b.Buildtemplate,
		Namespace:      b.Namespace,
		RegistrySecret: b.RegistrySecret,
	}

	sourceBt, err := bt.Get(clientset)
	if err != nil {
		cbt := clusterbuildtemplate.ClusterBuildtemplate{Name: bt.Name}
		sourceCbt, err := cbt.Get(clientset)
		if err != nil {
			return nil, nil
		}
		sourceBt.Spec = sourceCbt.Spec
		sourceBt.TypeMeta = sourceCbt.TypeMeta
		sourceBt.ObjectMeta = sourceCbt.ObjectMeta
	}
	return bt.Clone(*sourceBt, clientset)
}

// SetOwner sets passed owner object as owner reference in Build structure
func (b *Build) SetOwner(clientset *client.ConfigSet, owner metav1.OwnerReference) error {
	build, err := clientset.Build.BuildV1alpha1().Builds(b.Namespace).Get(b.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	build.SetOwnerReferences([]metav1.OwnerReference{owner})
	_, err = clientset.Build.BuildV1alpha1().Builds(b.Namespace).Update(build)
	return err
}

func (b *Build) buildSpecSourceURL() buildv1alpha1.BuildSpec {
	return buildv1alpha1.BuildSpec{
		Source: &buildv1alpha1.SourceSpec{
			Git: &buildv1alpha1.GitSourceSpec{
				Url:      b.Source,
				Revision: b.Revision,
			},
		},
	}
}

func (b *Build) buildSpecLocalPath() buildv1alpha1.BuildSpec {
	return buildv1alpha1.BuildSpec{
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
					uploadDoneTrigger, path.Base(b.Source))},
			},
		},
	}
}

func getBuildArguments(image, tag string, buildArgs []string) []buildv1alpha1.ArgumentSpec {
	args := []buildv1alpha1.ArgumentSpec{
		{
			Name:  "IMAGE",
			Value: image,
		}, {
			Name:  "TAG",
			Value: tag,
		},
	}
	for k, v := range mapFromSlice(buildArgs) {
		args = append(args, buildv1alpha1.ArgumentSpec{
			Name: k, Value: v,
		})
	}
	return args
}

func (b *Build) imageName(clientset *client.ConfigSet) (string, error) {
	if len(b.RegistrySecret) == 0 {
		return fmt.Sprintf("%s/%s/%s", b.Registry, b.Namespace, b.Name), nil
	}
	secret, err := clientset.Core.CoreV1().Secrets(b.Namespace).Get(b.RegistrySecret, metav1.GetOptions{})
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
			return fmt.Sprintf("%s/%s", url, b.Name), nil
		}
		return fmt.Sprintf("%s/%s/%s", k, v.Username, b.Name), nil
	}
	return "", errors.New("empty registry credentials")
}

func gitlabEnv() (string, bool) {
	return os.LookupEnv("CI_REGISTRY_IMAGE")
}

func (b *Build) wait(build *buildv1alpha1.Build, clientset *client.ConfigSet) error {
	watch, err := clientset.Build.BuildV1alpha1().Builds(b.Namespace).Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + build.Name,
	})
	if err != nil || watch == nil {
		return fmt.Errorf("Can't get build %q watch interface: %s", b.Name, err)
	}
	defer watch.Stop()

	duration, err := time.ParseDuration(b.Timeout)
	if err != nil {
		duration = 10 * time.Minute
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case event := <-watch.ResultChan():
			if event.Object == nil {
				return b.wait(build, clientset)
			}
			res, ok := event.Object.(*buildv1alpha1.Build)
			if !ok {
				continue
			}
			cond := res.Status.GetCondition(buildv1alpha1.BuildSucceeded)
			if cond == nil {
				// return fmt.Errorf("Can't get build %q condition", b.Name)
				continue
			}
			if cond.Status == corev1.ConditionUnknown {
				// Build is in progress
				continue
			}
			if cond.Status == corev1.ConditionTrue {
				// Build succeeded
				return nil
			}
			// Build failure
			return fmt.Errorf("Unexpected build status, please run \"tm -n %s get build %s\" to see details", b.Namespace, b.Name)
		case <-ticker.C:
			return fmt.Errorf("Watch build timeout")
		}
	}
	return nil
}
