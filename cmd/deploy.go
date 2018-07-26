package cmd

import (
	"errors"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	image  string
	source string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy knative service",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("Exactly one argument expected")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		deploy(args)
	},
}

func init() {
	deployCmd.Flags().StringVar(&image, "from-image", "", "Image to deploy")
	deployCmd.Flags().StringVar(&source, "from-source", "", "Source URL to deploy")
	rootCmd.AddCommand(deployCmd)
}

func deploy(args []string) {
	if len(image) != 0 && len(source) != 0 {
		log.Errorln("Only one image source allowed for deploy")
		return
	}

	spec := v1alpha1.ServiceSpec{}
	if len(image) != 0 {
		spec = v1alpha1.ServiceSpec{
			RunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					RevisionTemplate: v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							Container: corev1.Container{
								Image: image,
							},
						},
					},
				},
			},
		}
	} else if len(source) != 0 {
		spec = v1alpha1.ServiceSpec{
			RunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
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
					RevisionTemplate: v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
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

	s := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: namespace,
			Labels: map[string]string{
				"created-by": "tm",
			},
			// Annotations: map[string]string{},
		},

		Spec: spec,
	}

	service, err := serving.ServingV1alpha1().Services(namespace).Create(&s)
	if err != nil {
		log.Errorln(err)
		return
	}
	log.Debugf("+%v\n", service.Status)
	log.Infof("Service %s successfully deployed\n", service.Name)
}
