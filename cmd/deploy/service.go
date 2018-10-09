/*
Copyright (c) 2018 TriggerMesh, Inc

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

package deploy

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/triggermesh/tm/pkg/client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	image, source, url, storage, pullPolicy,
	memory, path, cpu, revision, buildtemplate,
	sourceDir, imageTag, handler string
	env, labels, secrets, buildArgs []string
)

func cmdDeployService(clientset *client.ClientSet) *cobra.Command {
	deployServiceCmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Deploy knative service",
		Args:    cobra.ExactArgs(1),
		Example: "tm -n default deploy service foo --build-template kaniko --build-argument SKIP_TLS_VERIFY=true --from-image gcr.io/google-samples/hello-app:1.0",
		Run: func(cmd *cobra.Command, args []string) {
			if err := deployService(args, clientset); err != nil {
				log.Fatalln(err)
			}
		},
	}

	deployServiceCmd.Flags().StringVar(&image, "from-image", "", "Image to deploy")
	deployServiceCmd.Flags().StringVar(&source, "from-source", "", "Git source URL to deploy")
	deployServiceCmd.Flags().StringVar(&revision, "revision", "master", "May be used with \"--from-source\" flag: git revision (branch, tag, commit SHA or ref) to clone")
	deployServiceCmd.Flags().StringVar(&path, "from-file", "", "Local file path to deploy")
	deployServiceCmd.Flags().StringVar(&url, "from-url", "", "File source URL to deploy")
	deployServiceCmd.Flags().StringVar(&buildtemplate, "build-template", "kaniko", "Build template to use with service")
	deployServiceCmd.Flags().StringVar(&sourceDir, "source-directory", "", "Workspace context directory to work in")
	deployServiceCmd.Flags().StringVar(&imageTag, "tag", "latest", "Image tag to build")
	deployServiceCmd.Flags().StringVar(&pullPolicy, "image-pull-policy", "Always", "Image pull policy")
	deployServiceCmd.Flags().StringSliceVar(&buildArgs, "build-argument", []string{}, "Image tag to build")
	deployServiceCmd.Flags().StringSliceVarP(&labels, "label", "l", []string{}, "Service labels")
	deployServiceCmd.Flags().StringSliceVarP(&env, "env", "e", []string{}, "Environment variables of the service, eg. `--env foo=bar`")
	// deployCmd.Flags().StringSliceVar(&secrets, "secrets", []string{}, "Name of Secrets to mount into service environment")
	// deployCmd.Flags().Int32Var(&port, "port", 8080, "Custom service port")

	return deployServiceCmd
}

func deployService(args []string, clientset *client.ClientSet) error {
	configuration := servingv1alpha1.ConfigurationSpec{}
	buildArguments, templateParams := getBuildArguments(fmt.Sprintf("%s/%s-%s-source", clientset.Registry, clientset.Namespace, args[0]), buildArgs)

	switch {
	case len(image) != 0:
		configuration = fromImage(args)
	case len(source) != 0:
		if err := createConfigMap(nil, clientset); err != nil {
			return err
		}
		configuration = fromSource(args, clientset.Registry, clientset.Namespace)
		if err := updateBuildTemplate(buildtemplate, templateParams, clientset); err != nil {
			return err
		}

		configuration.Build.Template = &buildv1alpha1.TemplateInstantiationSpec{
			Name:      buildtemplate,
			Arguments: buildArguments,
		}
	case len(url) != 0:
		configuration = fromURL(args, clientset.Registry, clientset.Namespace)
	case len(path) != 0:
		filebody, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		data := make(map[string]string)
		data[args[0]] = string(filebody)
		if err := createConfigMap(data, clientset); err != nil {
			return err
		}
		configuration = fromFile(args, clientset.Registry, clientset.Namespace)
	}

	envVars := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "timestamp",
			Value: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	for k, v := range getArgsFromSlice(env) {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}
	configuration.RevisionTemplate.Spec.Container.Env = envVars
	configuration.RevisionTemplate.Spec.Container.ImagePullPolicy = corev1.PullPolicy(pullPolicy)
	serviceObject := servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/servingv1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: clientset.Namespace,
			CreationTimestamp: metav1.Time{
				time.Now(),
			},
			Labels: getArgsFromSlice(labels),
		},

		Spec: servingv1alpha1.ServiceSpec{
			RunLatest: &servingv1alpha1.RunLatestType{
				Configuration: configuration,
			},
		},
	}

	service, err := clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).Get(args[0], metav1.GetOptions{})
	if err == nil {
		serviceObject.ObjectMeta.ResourceVersion = service.ObjectMeta.ResourceVersion
		service, err = clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).Update(&serviceObject)
		if err != nil {
			return err
		}
		fmt.Println("Service update started. Run \"tm -n %s get revisions\" to see available revisions", clientset.Namespace)
	} else if k8sErrors.IsNotFound(err) {
		service, err := clientset.Serving.ServingV1alpha1().Services(clientset.Namespace).Create(&serviceObject)
		if err != nil {
			return err
		}
		fmt.Println("Deployment started. Run \"tm -n %s describe service %s\" to see the details", clientset.Namespace, service.Name)
	} else {
		return err
	}
	return nil
}

func fromImage(args []string) servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "true",
				},
				Name: args[0],
			},
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: image,
				},
			},
		},
	}
}

func fromSource(args []string, registry, namespace string) servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Git: &buildv1alpha1.GitSourceSpec{
					Url:      source,
					Revision: revision,
				},
			},
		},
		RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "true",
				},
				Name: args[0],
			},
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: fmt.Sprintf("%s/%s-%s-source:%s", registry, namespace, args[0], imageTag),
				},
			},
		},
	}
}

func fromURL(args []string, registry, namespace string) servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Custom: &corev1.Container{
					Image: "registry.hub.docker.com/library/busybox",
				},
			},
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name: "getandbuild",
			},
		},
		RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "true",
				},
				Name: args[0],
			},
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: fmt.Sprintf("%s/%s-%s-url:%s", registry, namespace, args[0], imageTag),
				},
			},
		},
	}
}

func fromFile(args []string, registry, namespace string) servingv1alpha1.ConfigurationSpec {
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Custom: &corev1.Container{
					Image: "registry.hub.docker.com/library/busybox",
				},
			},
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name: "kaniko",
			},
		},
		RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: args[0],
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "true",
				},
			},
			Spec: servingv1alpha1.RevisionSpec{
				Container: corev1.Container{
					Image: fmt.Sprintf("%s/%s-%s-file:%s", registry, namespace, args[0], imageTag),
				},
			},
		},
	}
}

func createConfigMap(data map[string]string, clientset *client.ClientSet) error {
	newmap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dockerfile",
			Namespace: clientset.Namespace,
		},
		Data: data,
	}
	cm, err := clientset.Core.CoreV1().ConfigMaps(clientset.Namespace).Get("dockerfile", metav1.GetOptions{})
	if err == nil {
		newmap.ObjectMeta.ResourceVersion = cm.ObjectMeta.ResourceVersion
		_, err = clientset.Core.CoreV1().ConfigMaps(clientset.Namespace).Update(&newmap)
		return err
	} else if k8sErrors.IsNotFound(err) {
		_, err = clientset.Core.CoreV1().ConfigMaps(clientset.Namespace).Create(&newmap)
		return err
	}
	return err
}

func getArgsFromSlice(slice []string) map[string]string {
	m := make(map[string]string)
	for _, s := range slice {
		t := regexp.MustCompile("[:=]").Split(s, 2)
		if len(t) != 2 {
			fmt.Println("Can't parse argument slice %s", s)
			continue
		}
		m[t[0]] = t[1]
	}
	return m
}

func updateBuildTemplate(name string, params []buildv1alpha1.ParameterSpec, clientset *client.ClientSet) error {
	buildTemplate, err := clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Matching new build parameters with existing to check if need to update build template
	var new bool
	for _, v := range params {
		new = true
		for _, vv := range buildTemplate.Spec.Parameters {
			if v.Name == vv.Name {
				new = false
				break
			}
		}
		if new {
			break
		}
	}

	if new {
		buildTemplate.Spec.Parameters = params
		_, err = clientset.Build.BuildV1alpha1().BuildTemplates(clientset.Namespace).Update(buildTemplate)
	}

	return err
}
