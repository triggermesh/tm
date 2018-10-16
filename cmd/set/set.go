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

package set

import (
	"fmt"
	"regexp"
	"strconv"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/triggermesh/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Route struct {
	Revisions []string
	Configs   []string
}

func split(slice []string) map[string]int {
	m := make(map[string]int)
	for _, s := range slice {
		t := regexp.MustCompile("[:=]").Split(s, 2)
		if len(t) != 2 {
			fmt.Printf("Can't parse target %s\n", s)
			continue
		}
		percent, err := strconv.Atoi(t[1])
		if err != nil {
			fmt.Printf("Invalid traffic percent value %s\n", t[1])
			continue
		}
		m[t[0]] = percent
	}
	return m
}

func (r *Route) SetPercentage(args []string, clientset *client.ClientSet) error {
	targets := []servingv1alpha1.TrafficTarget{}
	// TODO: add named target support
	for revision, percent := range split(r.Revisions) {
		targets = append(targets, servingv1alpha1.TrafficTarget{
			RevisionName: revision,
			Percent:      percent,
		})
	}
	for config, percent := range split(r.Configs) {
		targets = append(targets, servingv1alpha1.TrafficTarget{
			ConfigurationName: config,
			Percent:           percent,
		})
	}

	routeOld, err := clientset.Serving.ServingV1alpha1().Routes(clientset.Namespace).Get(args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	route := servingv1alpha1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "serving.knative.dev/servingv1alpha1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      args[0],
			Namespace: clientset.Namespace,
			Labels: map[string]string{
				"created-by":                  "tm",
				"serving.knative.dev/service": args[0],
			},
		},
	}
	route.ObjectMeta.ResourceVersion = routeOld.ObjectMeta.ResourceVersion
	route.Spec.Traffic = targets
	if _, err = clientset.Serving.ServingV1alpha1().Routes(clientset.Namespace).Update(&route); err != nil {
		return err
	}
	return nil
}
