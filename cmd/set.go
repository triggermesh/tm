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

package cmd

import (
	"regexp"
	"strconv"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	revisions, configs []string
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource parameters",
}

var setRouteCmd = &cobra.Command{
	Use:   "route",
	Short: "Configure service route",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := setPercentage(args); err != nil {
			log.Errorln(err)
			return
		}
		log.Infoln("Routes successfully updated")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setRouteCmd)
	setRouteCmd.Flags().StringSliceVarP(&revisions, "revisions", "r", []string{}, "Set traffic percentage for revision")
	setRouteCmd.Flags().StringSliceVarP(&configs, "configs", "c", []string{}, "Set traffic percentage for configuration")
}

func split(slice []string) map[string]int {
	m := make(map[string]int)
	for _, s := range slice {
		t := regexp.MustCompile("[:=]").Split(s, 2)
		if len(t) != 2 {
			log.Warnf("Can't parse target %s", s)
			continue
		}
		percent, err := strconv.Atoi(t[1])
		if err != nil {
			log.Warnf("Invalid traffic percent value %s", t[1])
			continue
		}
		m[t[0]] = percent
	}
	return m
}

func setPercentage(args []string) error {
	targets := []servingv1alpha1.TrafficTarget{}
	// TODO: add named target support
	for revision, percent := range split(revisions) {
		targets = append(targets, servingv1alpha1.TrafficTarget{
			RevisionName: revision,
			Percent:      percent,
		})
	}
	for config, percent := range split(configs) {
		targets = append(targets, servingv1alpha1.TrafficTarget{
			ConfigurationName: config,
			Percent:           percent,
		})
	}

	r, err := serving.ServingV1alpha1().Routes(namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	// fmt.Printf("%+v\n", routes)
	route := servingv1alpha1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "serving.knative.dev/servingv1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: namespace,
			Labels: map[string]string{
				"created-by":                  "tm",
				"serving.knative.dev/service": args[0],
			},
		},
	}
	route.ObjectMeta.ResourceVersion = r.ObjectMeta.ResourceVersion
	route.Spec.Traffic = targets
	if _, err = serving.ServingV1alpha1().Routes(namespace).Update(&route); err != nil {
		return err
	}
	return nil
}
