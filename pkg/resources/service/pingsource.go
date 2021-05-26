// Copyright 2020 TriggerMesh Inc.
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
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"

	"github.com/triggermesh/tm/pkg/client"
)

const serviceLabelKey = "cli.triggermesh.io/service"

func (s *Service) pingSource(schedule, data string, owner kmeta.OwnerRefable) *sourcesv1.PingSource {
	return &sourcesv1.PingSource{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: s.Name + "-",
			Namespace:    s.Namespace,
			Labels: map[string]string{
				serviceLabelKey: s.Name,
			},
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(owner)},
		},
		Spec: sourcesv1.PingSourceSpec{
			Schedule: schedule,
			Data:     data,
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						APIVersion: owner.GetGroupVersionKind().Version,
						Kind:       owner.GetGroupVersionKind().Kind,
						Name:       owner.GetObjectMeta().GetName(),
						Namespace:  owner.GetObjectMeta().GetNamespace(),
					},
				},
			},
		},
	}
}

func (s *Service) createPingSource(ps *sourcesv1.PingSource, clientset *client.ConfigSet) error {
	_, err := clientset.Eventing.SourcesV1().PingSources(ps.Namespace).Create(context.Background(), ps, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create PingSource %q: %w", ps.Name, err)
	}
	return nil
}

func (s *Service) removePingSources(uid types.UID, clientset *client.ConfigSet) error {
	err := clientset.Eventing.SourcesV1().PingSources(s.Namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: serviceLabelKey + "=" + s.Name,
	})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot remove owned PingSources: %w", err)
	}
	return nil
}
