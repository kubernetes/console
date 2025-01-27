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

package pod

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	metricapi "k8s.io/dashboard/api/pkg/integration/metric/api"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/api/pkg/resource/event"
	"k8s.io/dashboard/types"
)

// getRestartCount return the restart count of given pod (total number of its containers restarts).
func getRestartCount(pod v1.Pod) int32 {
	var restartCount int32 = 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	return restartCount
}

// getPodStatus returns status string calculated based on the same logic as kubectl
// Base code: https://github.com/kubernetes/kubernetes/blob/v1.31.1/pkg/printers/internalversion/printers.go
//
// nolint:gocyclo // disable cyclo check as this function is copied from kubectl
func getPodStatus(pod v1.Pod) string {
	podPhase := pod.Status.Phase
	reason := string(podPhase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	// If the Pod carries {type:PodScheduled, reason:SchedulingGated}, set reason to 'SchedulingGated'.
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Reason == v1.PodReasonSchedulingGated {
			reason = v1.PodReasonSchedulingGated
		}
	}

	initContainers := make(map[string]*v1.Container)
	for i := range pod.Spec.InitContainers {
		initContainers[pod.Spec.InitContainers[i].Name] = &pod.Spec.InitContainers[i]
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case isRestartableInitContainer(initContainers[container.Name]) &&
			container.Started != nil && *container.Started:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init: Signal %d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init: ExitCode %d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = fmt.Sprintf("Init: %s", container.State.Waiting.Reason)
			initializing = true
		default:
			reason = fmt.Sprintf("Init: %d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}

	if !initializing || isPodInitializedConditionTrue(&pod.Status) {
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal: %d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode: %d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = string(v1.PodRunning)
			} else {
				reason = "NotReady"
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == types.NodeUnreachablePodReason {
		reason = string(v1.PodUnknown)
	} else if pod.DeletionTimestamp != nil && !IsPodPhaseTerminal(podPhase) {
		reason = "Terminating"
	}

	return reason
}

func hasPodReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// getPodStatusPhase returns one of four pod status phases (Pending, Running, Succeeded, Failed, Unknown, Terminating)
func getPodStatusPhase(pod v1.Pod, warnings []common.Event) v1.PodPhase {
	// For terminated pods that failed
	if pod.Status.Phase == v1.PodFailed {
		return v1.PodFailed
	}

	// For successfully terminated pods
	if pod.Status.Phase == v1.PodSucceeded {
		return v1.PodSucceeded
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

	if initialized && ready && pod.Status.Phase == v1.PodRunning {
		return v1.PodRunning
	}

	// If the pod would otherwise be pending but has warning then label it as
	// failed and show and error to the user.
	if len(warnings) > 0 {
		return v1.PodFailed
	}

	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	return ""
}

// The code below allows to perform complex data section on []api.Pod

type PodCell v1.Pod

func (self PodCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.StatusProperty:
		return dataselect.StdComparableString(getPodStatus(v1.Pod(self)))
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

func (self PodCell) GetResourceSelector() *metricapi.ResourceSelector {
	return &metricapi.ResourceSelector{
		Namespace:    self.ObjectMeta.Namespace,
		ResourceType: types.ResourceKindPod,
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

func getStatus(list *v1.PodList, events []v1.Event) common.ResourceStatus {
	info := common.ResourceStatus{}
	if list == nil {
		return info
	}

	for _, pod := range list.Items {
		warnings := event.GetPodsEventWarnings(events, []v1.Pod{pod})
		switch getPodStatusPhase(pod, warnings) {
		case v1.PodFailed:
			info.Failed++
		case v1.PodSucceeded:
			info.Succeeded++
		case v1.PodRunning:
			info.Running++
		case v1.PodPending:
			info.Pending++
		case "Terminating":
			info.Terminating++
		}
	}

	return info
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}

// PodRequestsAndLimits returns a dictionary of all defined resources summed up for all
// containers of the pod. If pod overhead is non-nil, the pod overhead is added to the
// total container resource requests and to the total container limits which have a
// non-zero quantity.
func PodRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}

	// Add overhead for running a pod to the sum of requests and to non-zero limits:
	if pod.Spec.Overhead != nil {
		addResourceList(reqs, pod.Spec.Overhead)

		for name, quantity := range pod.Spec.Overhead {
			if value, ok := limits[name]; ok && !value.IsZero() {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	return
}

func getPodAllocatedResources(pod *v1.Pod) (PodAllocatedResources, error) {
	reqs, limits, err := PodRequestsAndLimits(pod)
	if err != nil {
		return PodAllocatedResources{}, err
	}

	for podReqName, podReqValue := range reqs {
		if value, ok := reqs[podReqName]; !ok {
			reqs[podReqName] = podReqValue.DeepCopy()
		} else {
			value.Add(podReqValue)
			reqs[podReqName] = value
		}
	}

	for podLimitName, podLimitValue := range limits {
		if value, ok := limits[podLimitName]; !ok {
			limits[podLimitName] = podLimitValue.DeepCopy()
		} else {
			value.Add(podLimitValue)
			limits[podLimitName] = value
		}
	}

	cpuRequests, cpuLimits, memoryRequests, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]

	return PodAllocatedResources{
		CPURequests:    cpuRequests.MilliValue(),
		CPULimits:      cpuLimits.MilliValue(),
		MemoryRequests: memoryRequests.Value(),
		MemoryLimits:   memoryLimits.Value(),
	}, nil
}

func isPodInitializedConditionTrue(status *v1.PodStatus) bool {
	for _, condition := range status.Conditions {
		if condition.Type != v1.PodInitialized {
			continue
		}

		return condition.Status == v1.ConditionTrue
	}
	return false
}

func isRestartableInitContainer(initContainer *v1.Container) bool {
	if initContainer == nil {
		return false
	}
	if initContainer.RestartPolicy == nil {
		return false
	}

	return *initContainer.RestartPolicy == v1.ContainerRestartPolicyAlways
}

// IsPodPhaseTerminal returns true if the pod's phase is terminal.
func IsPodPhaseTerminal(phase v1.PodPhase) bool {
	return phase == v1.PodFailed || phase == v1.PodSucceeded
}
