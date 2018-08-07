package cmd

import (
	"fmt"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	image, source, ports string
	df                   = "/workspace/Dockerfile"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy knative service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// prepareNetwork(args)
		deployService(args)
	},
}

func init() {
	deployCmd.Flags().StringVar(&image, "from-image", "", "Image to deploy")
	deployCmd.Flags().StringVar(&source, "from-source", "", "Source URL to deploy")
	rootCmd.AddCommand(deployCmd)
}

func prepareNetwork(args []string) {
	// net := istiov1alpha3.VirtualService{}

	vs, err := serving.NetworkingV1alpha3().Gateways(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln(err)
		return
	}
	fmt.Println(vs.Items)
}

func deployService(args []string) {
	if len(image) != 0 && len(source) != 0 {
		log.Errorln("Only one image source allowed for deploy")
		return
	}

	configuration := servingv1alpha1.ConfigurationSpec{}
	if len(image) != 0 {
		configuration = fromImage(args)
	} else if len(source) != 0 {
		if err := deployKaniko(); err != nil {
			log.Errorln(err)
			return
		}
		configuration = fromSource(args)
	}

	s := servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/servingv1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: namespace,
			Labels: map[string]string{
				"created-by": "tm",
				// "knative":    "ingressgateway",
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

	service, err := serving.ServingV1alpha1().Services(namespace).Create(&s)
	if err != nil {
		log.Errorln(err)
		return
	}
	log.Infof("Service %s successfully deployed\n", service.Name)
}

func deployKaniko() error {
	_, err := build.BuildV1alpha1().BuildTemplates(namespace).Get("kaniko", metav1.GetOptions{})
	if err == nil {
		log.Debugln("kaniko already exist")
		return nil
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
					buildv1alpha1.ParameterSpec{
						Name: "IMAGE",
					},
					buildv1alpha1.ParameterSpec{
						Name:    "DOCKERFILE",
						Default: &df,
					},
				},
				Steps: []corev1.Container{
					corev1.Container{
						Name:  "build-and-push",
						Image: "gcr.io/kaniko-project/executor",
						Args:  []string{"--dockerfile=${DOCKERFILE}", "--destination=${IMAGE}"},
						VolumeMounts: []corev1.VolumeMount{
							corev1.VolumeMount{
								Name:      "docker-secret",
								MountPath: "/secret",
							},
						},
						Env: []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "DOCKER_APPLICATION_CREDENTIALS",
								Value: "/secret/docker-secret.json",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					corev1.Volume{
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
				// Labels: map[string]string{
				// "knative": "ingressgateway",
				// },
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
						Name:  "IMAGE",
						Value: "docker.io/triggermesh/" + args[0] + "-from-source:latest",
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
