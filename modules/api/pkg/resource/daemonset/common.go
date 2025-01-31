// Copyright 2017 The Kubernetes Authors.
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

package daemonset

import (
	"context"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	client "k8s.io/client-go/kubernetes"

	metricapi "k8s.io/dashboard/api/pkg/integration/metric/api"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/api/pkg/resource/event"
	"k8s.io/dashboard/types"
)

// GetServicesForDSDeletion is based on given selector returns list of services that are candidates for deletion.
// Services are matched by daemon sets' label selector. They are deleted if given
// label selector is targeting only 1 daemon set.
func GetServicesForDSDeletion(client client.Interface, labelSelector labels.Selector,
	namespace string) ([]v1.Service, error) {

	daemonSet, err := client.AppsV1().DaemonSets(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: labelSelector.String(),
		FieldSelector: fields.Everything().String(),
	})
	if err != nil {
		return nil, err
	}

	// if label selector is targeting only 1 daemon set
	// then we can delete services targeted by this label selector,
	// otherwise we can not delete any services so just return empty list
	if len(daemonSet.Items) != 1 {
		return []v1.Service{}, nil
	}

	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: labelSelector.String(),
		FieldSelector: fields.Everything().String(),
	})
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}

// The code below allows to perform complex data section on Daemon Set

type DaemonSetCell apps.DaemonSet

func (self DaemonSetCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	case dataselect.LabelProperty:
		return dataselect.StdComparableMap(self.ObjectMeta.Labels)
	case dataselect.AnnotationsProperty:
		return dataselect.StdComparableMap(self.ObjectMeta.Annotations)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func (self DaemonSetCell) GetResourceSelector() *metricapi.ResourceSelector {
	return &metricapi.ResourceSelector{
		Namespace:    self.ObjectMeta.Namespace,
		ResourceType: types.ResourceKindDaemonSet,
		ResourceName: self.ObjectMeta.Name,
		UID:          self.UID,
	}
}

func ToCells(std []apps.DaemonSet) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = DaemonSetCell(std[i])
	}
	return cells
}

func FromCells(cells []dataselect.DataCell) []apps.DaemonSet {
	std := make([]apps.DaemonSet, len(cells))
	for i := range std {
		std[i] = apps.DaemonSet(cells[i].(DaemonSetCell))
	}
	return std
}

func getStatus(list *apps.DaemonSetList, pods []v1.Pod, events []v1.Event) common.ResourceStatus {
	info := common.ResourceStatus{}
	if list == nil {
		return info
	}

	for _, daemonSet := range list.Items {
		matchingPods := common.FilterPodsByControllerRef(&daemonSet, pods)
		podInfo := common.GetPodInfo(daemonSet.Status.CurrentNumberScheduled,
			&daemonSet.Status.DesiredNumberScheduled, matchingPods)
		warnings := event.GetPodsEventWarnings(events, matchingPods)

		if len(warnings) > 0 {
			info.Failed++
		} else if podInfo.Pending > 0 {
			info.Pending++
		} else {
			info.Running++
		}
	}

	return info
}
