package cmd

import (
	"fmt"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	image, source, ports string
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
	vs, err := serving.NetworkingV1alpha3().VirtualServices(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		log.Errorln(err)
		return
	}
	fmt.Println(vs.Spec.Hosts)
}

func deployService(args []string) {
	if len(image) != 0 && len(source) != 0 {
		log.Errorln("Only one image source allowed for deploy")
		return
	}

	spec := servingv1alpha1.ServiceSpec{}
	if len(image) != 0 {
		spec = servingv1alpha1.ServiceSpec{
			RunLatest: &servingv1alpha1.RunLatestType{
				Configuration: servingv1alpha1.ConfigurationSpec{
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
				},
			},
		}
	} else if len(source) != 0 {
		spec = servingv1alpha1.ServiceSpec{
			RunLatest: &servingv1alpha1.RunLatestType{
				Configuration: servingv1alpha1.ConfigurationSpec{
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
									Value: "&image docker.io/{DOCKER_USERNAME}/app-from-source:latest",
								},
							},
						},
					},
					RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
						Spec: servingv1alpha1.RevisionSpec{
							Container: corev1.Container{
								Image:           image,
								ImagePullPolicy: corev1.PullAlways,
							},
						},
					},
				},
			},
		}
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
			},
			// Annotations: map[string]string{
			// "test": "test",
			// },
		},

		Spec: spec,
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
