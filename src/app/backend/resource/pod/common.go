// Copyright 2017 The Kubernetes Dashboard Authors.
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

package pod

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"k8s.io/client-go/pkg/api/v1"
)

// Gets restart count of given pod (total number of its containers restarts).
func getRestartCount(pod v1.Pod) int32 {
	var restartCount int32 = 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	return restartCount
}

// getPodStatus returns a PodStatus object containing a summary of the pod's status.
func getPodStatus(pod v1.Pod, warnings []common.Event) PodStatus {
	var states []v1.ContainerState
	for _, containerStatus := range pod.Status.ContainerStatuses {
		states = append(states, containerStatus.State)
	}

	return PodStatus{
		Status:          getPodStatusStatus(pod, warnings),
		PodPhase:        pod.Status.Phase,
		ContainerStates: states,
	}
}

// getPodStatus returns one of three pod statuses (pending, success, failed)
func getPodStatusStatus(pod v1.Pod, warnings []common.Event) string {
	// For terminated pods that failed
	if pod.Status.Phase == v1.PodFailed {
		return "failed"
	}

	// For successfully terminated pods
	if pod.Status.Phase == v1.PodSucceeded {
		return "success"
	}

	ready := false
	initialized := false
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady {
			ready = c.Status == v1.ConditionTrue
		}
		if c.Type == v1.PodInitialized {
			initialized = c.Status == v1.ConditionTrue
		}
	}

	if initialized && ready {
		return "success"
	}

	// If the pod would otherwise be pending but has warning then label it as
	// failed and show and error to the user.
	if len(warnings) > 0 {
		return "failed"
	}

	// Unknown?
	return "pending"
}

// ToPod transforms Kubernetes pod object into object returned by API.
func ToPod(pod *v1.Pod, metrics *MetricsByPod, warnings []common.Event) Pod {
	podDetail := Pod{
		ObjectMeta:   api.NewObjectMeta(pod.ObjectMeta),
		TypeMeta:     api.NewTypeMeta(api.ResourceKindPod),
		PodStatus:    getPodStatus(*pod, warnings),
		RestartCount: getRestartCount(*pod),
	}

	if m, exists := metrics.MetricsMap[pod.UID]; exists {
		podDetail.Metrics = &m
	}

	return podDetail
}

// The code below allows to perform complex data section on []api.Pod

type PodCell v1.Pod

func (self PodCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	case dataselect.StatusProperty:
		return dataselect.StdComparableString(self.Status.Phase)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func (self PodCell) GetResourceSelector() *metricapi.ResourceSelector {
	return &metricapi.ResourceSelector{
		Namespace:    self.ObjectMeta.Namespace,
		ResourceType: api.ResourceKindPod,
		ResourceName: self.ObjectMeta.Name,
		UID:          self.ObjectMeta.UID,
	}
}

func toCells(std []v1.Pod) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = PodCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []v1.Pod {
	std := make([]v1.Pod, len(cells))
	for i := range std {
		std[i] = v1.Pod(cells[i].(PodCell))
	}
	return std
}

func getPodConditions(pod v1.Pod) []common.Condition {
	var conditions []common.Condition
	for _, condition := range pod.Status.Conditions {
		conditions = append(conditions, common.Condition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastProbeTime:      condition.LastProbeTime,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}
	return conditions
}
