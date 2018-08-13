package cmd

import (
	"io/ioutil"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	image, source, url, path string
	df                       = "/workspace/Dockerfile"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy knative service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deployService(args); err != nil {
			log.Errorln(err)
		}
	},
}

func init() {
	deployCmd.Flags().StringVar(&image, "from-image", "", "Image to deploy")
	deployCmd.Flags().StringVar(&source, "from-source", "", "Git source URL to deploy")
	deployCmd.Flags().StringVar(&path, "from-file", "", "Local file path to deploy")
	deployCmd.Flags().StringVar(&url, "from-url", "", "File source URL to deploy")
	rootCmd.AddCommand(deployCmd)
}

func deployService(args []string) error {
	configuration := servingv1alpha1.ConfigurationSpec{}
	switch {
	case len(image) != 0:
		configuration = fromImage(args)
	case len(source) != 0:
		if err := kanikoBuildTemplate(); err != nil {
			return err
		}
		configuration = fromSource(args)
	case len(url) != 0:
		if err := getterBuildTemplate(); err != nil {
			return err
		}
		configuration = fromURL(args)
	case len(path) != 0:
		if err := createConfigMap(args); err != nil {
			return err
		}
		if err := kanikoBuildTemplate(); err != nil {
			return err
		}
		configuration = fromFile(args)
	}

	configuration.RevisionTemplate.Spec.Container.Env = []corev1.EnvVar{
		{
			Name:  "timestamp",
			Value: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	s := servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/servingv1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: namespace,
			CreationTimestamp: metav1.Time{
				time.Now(),
			},
			Labels: map[string]string{
				"created-by": "tm",
			},
		},

		Spec: servingv1alpha1.ServiceSpec{
			RunLatest: &servingv1alpha1.RunLatestType{
				Configuration: configuration,
			},
		},
	}

	log.Debugf("Service object: %+v\n", s)
	log.Debugf("Service specs: %+v\n", s.Spec.RunLatest)

	service, err := serving.ServingV1alpha1().Services(namespace).Get(args[0], metav1.GetOptions{})
	if err == nil {
		s.ObjectMeta.ResourceVersion = service.ObjectMeta.ResourceVersion
		service, err = serving.ServingV1alpha1().Services(namespace).Update(&s)
		if err != nil {
			return err
		}
		log.Infof("Service update started. Run \"tm -n %s get revisions %s\" to see available revisions\n", namespace, service.Name)
	} else if k8sErrors.IsNotFound(err) {
		service, err := serving.ServingV1alpha1().Services(namespace).Create(&s)
		if err != nil {
			return err
		}
		log.Infof("Deployment started. Run \"tm -n %s describe service %s\" to see the details\n", namespace, service.Name)
	} else {
		return err
	}
	return nil
}

func kanikoBuildTemplate() error {
	_, err := build.BuildV1alpha1().BuildTemplates(namespace).Get("kaniko", metav1.GetOptions{})
	if err == nil {
		log.Debugln("kaniko already exist")
	} else if k8sErrors.IsNotFound(err) {
		log.Debugln("deploying kaniko")
		bt := buildv1alpha1.BuildTemplate{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BuildTemplate",
				APIVersion: "build.knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kaniko",
				Namespace: namespace,
			},
			Spec: buildv1alpha1.BuildTemplateSpec{
				Parameters: []buildv1alpha1.ParameterSpec{
					{
						Name: "IMAGE",
					},
					{
						Name:    "DOCKERFILE",
						Default: &df,
					},
				},
				Steps: []corev1.Container{
					{
						Name:  "build-and-push",
						Image: "gcr.io/kaniko-project/executor",
						Args:  []string{"--dockerfile=${DOCKERFILE}", "--destination=${IMAGE}"},
						Env: []corev1.EnvVar{
							{
								Name:  "DOCKER_CONFIG",
								Value: "/docker-config",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "docker-secret",
								MountPath: "/docker-config",
							},
							{
								Name:      "docker-file",
								MountPath: "/docker-file",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "docker-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "docker-secret",
							},
						},
					},
					{
						Name: "docker-file",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "dockerfile"},
							},
						},
					},
				},
			},
		}
		_, err = build.BuildV1alpha1().BuildTemplates(namespace).Create(&bt)
		return err
	}
	return err
}

func getterBuildTemplate() error {
	_, err := build.BuildV1alpha1().BuildTemplates(namespace).Get("getandbuild", metav1.GetOptions{})
	if err == nil {
		log.Debugln("getandbuild template already exist")
	} else if k8sErrors.IsNotFound(err) {
		log.Debugln("deploying getandbuild template")
		bt := buildv1alpha1.BuildTemplate{
			TypeMeta: metav1.TypeMeta{
				Kind:       "BuildTemplate",
				APIVersion: "build.knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "getandbuild",
				Namespace: namespace,
			},
			Spec: buildv1alpha1.BuildTemplateSpec{
				Parameters: []buildv1alpha1.ParameterSpec{
					{
						Name: "URL",
					},
					{
						Name: "IMAGE",
					},
					{
						Name:    "DOCKERFILE",
						Default: &df,
					},
				},
				Steps: []corev1.Container{
					{
						Name:  "get",
						Image: "index.docker.io/byrnedo/alpine-curl",
						Args:  []string{"-o", "${DOCKERFILE}", "${URL}"},
					},
					{
						Name:  "build-and-push",
						Image: "gcr.io/kaniko-project/executor",
						Args:  []string{"--dockerfile=${DOCKERFILE}", "--destination=${IMAGE}"},
						Env: []corev1.EnvVar{
							{
								Name:  "DOCKER_CONFIG",
								Value: "/docker-config",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "docker-secret",
								MountPath: "/docker-config",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "docker-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "docker-secret",
							},
						},
					},
				},
			},
		}
		_, err = build.BuildV1alpha1().BuildTemplates(namespace).Create(&bt)
		return err
	}
	return err
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

func fromSource(args []string) servingv1alpha1.ConfigurationSpec {
	image = "index.docker.io/triggermesh/" + args[0] + "-from-source:latest"
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Git: &buildv1alpha1.GitSourceSpec{
					Url:      source,
					Revision: "master",
				},
			},
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name: "kaniko",
				Arguments: []buildv1alpha1.ArgumentSpec{
					{
						Name: "IMAGE",
						// TODO: replace triggermesh docker registry account
						Value: image,
					},
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
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
	}
}

func fromURL(args []string) servingv1alpha1.ConfigurationSpec {
	image = "index.docker.io/triggermesh/" + args[0] + "-from-url:latest"
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Custom: &corev1.Container{
					Image: "registry.hub.docker.com/library/busybox",
				},
			},
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name: "getandbuild",
				Arguments: []buildv1alpha1.ArgumentSpec{
					{
						Name: "IMAGE",
						// TODO: replace triggermesh docker registry account
						Value: image,
					},
					{
						Name:  "URL",
						Value: url,
					},
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
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
	}
}

func fromFile(args []string) servingv1alpha1.ConfigurationSpec {
	image = "index.docker.io/triggermesh/" + args[0] + "-from-file:latest"
	return servingv1alpha1.ConfigurationSpec{
		Build: &buildv1alpha1.BuildSpec{
			Source: &buildv1alpha1.SourceSpec{
				Custom: &corev1.Container{
					Image: "registry.hub.docker.com/library/busybox",
				},
			},
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name: "kaniko",
				Arguments: []buildv1alpha1.ArgumentSpec{
					{
						Name: "IMAGE",
						// TODO: replace triggermesh docker registry account
						Value: image,
					},
					{
						Name:  "DOCKERFILE",
						Value: "/docker-file/" + args[0],
					},
				},
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
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
	}
}

func createConfigMap(args []string) error {
	filebody, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	newmap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dockerfile",
			Namespace: namespace,
		},
		Data: map[string]string{
			args[0]: string(filebody),
		},
	}
	cm, err := core.CoreV1().ConfigMaps(namespace).Get("dockerfile", metav1.GetOptions{})
	if err == nil {
		newmap.ObjectMeta.ResourceVersion = cm.ObjectMeta.ResourceVersion
		_, err = core.CoreV1().ConfigMaps(namespace).Update(&newmap)
		return err
	} else if k8sErrors.IsNotFound(err) {
		_, err = core.CoreV1().ConfigMaps(namespace).Create(&newmap)
		return err
	}
	return err
}
