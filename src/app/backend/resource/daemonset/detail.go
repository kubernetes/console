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

package daemonset

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	ds "github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	resourceService "github.com/kubernetes/dashboard/src/app/backend/resource/service"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

// DaemonSeDetail represents detailed information about a Daemon Set.
type DaemonSetDetail struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Label selector of the Daemon Set.
	LabelSelector *v1.LabelSelector `json:"labelSelector,omitempty"`

	// Container image list of the pod template specified by this Daemon Set.
	ContainerImages []string `json:"containerImages"`

	// Aggregate information about pods of this daemon set.
	PodInfo common.PodInfo `json:"podInfo"`

	// Detailed information about Pods belonging to this Daemon Set.
	PodList pod.PodList `json:"podList"`

	// Detailed information about service related to Daemon Set.
	ServiceList resourceService.ServiceList `json:"serviceList"`

	// True when the data contains at least one pod with metrics information, false otherwise.
	HasMetrics bool `json:"hasMetrics"`

	// List of events related to this daemon set
	EventList common.EventList `json:"eventList"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// Returns detailed information about the given daemon set in the given namespace.
func GetDaemonSetDetail(client k8sClient.Interface, metricClient metricapi.MetricClient,
	namespace, name string) (*DaemonSetDetail, error) {
	log.Printf("Getting details of %s daemon set in %s namespace", name, namespace)

	daemonSet, err := client.ExtensionsV1beta1().DaemonSets(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	podList, err := GetDaemonSetPods(client, metricClient, ds.DefaultDataSelectWithMetrics, name, namespace)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	podInfo, err := getDaemonSetPodInfo(client, daemonSet)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	serviceList, err := GetDaemonSetServices(client, ds.DefaultDataSelect, namespace, name)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	eventList, err := event.GetResourceEvents(client, ds.DefaultDataSelect, daemonSet.Namespace, daemonSet.Name)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	daemonSetDetail := &DaemonSetDetail{
		ObjectMeta:    api.NewObjectMeta(daemonSet.ObjectMeta),
		TypeMeta:      api.NewTypeMeta(api.ResourceKindDaemonSet),
		LabelSelector: daemonSet.Spec.Selector,
		PodInfo:       *podInfo,
		PodList:       *podList,
		ServiceList:   *serviceList,
		EventList:     *eventList,
		Errors:        nonCriticalErrors,
	}

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		daemonSetDetail.ContainerImages = append(daemonSetDetail.ContainerImages,
			container.Image)
	}

	return daemonSetDetail, nil
}

// TODO(floreks): This should be transactional to make sure that DS will not be deleted without pods
// Deletes daemon set with given name in given namespace and related pods.
// Also deletes services related to daemon set if deleteServices is true.
func DeleteDaemonSet(client k8sClient.Interface, namespace, name string,
	deleteServices bool) error {

	log.Printf("Deleting %s daemon set from %s namespace", name, namespace)

	if deleteServices {
		if err := DeleteDaemonSetServices(client, namespace, name); err != nil {
			return err
		}
	}

	pods, err := getRawDaemonSetPods(client, namespace, name)
	if err != nil {
		return err
	}

	if err := client.Extensions().DaemonSets(namespace).Delete(name, &metaV1.DeleteOptions{}); err != nil {
		return err
	}

	for _, pod := range pods {
		if err := client.Core().Pods(namespace).Delete(pod.Name, &metaV1.DeleteOptions{}); err != nil {
			return err
		}
	}

	log.Printf("Successfully deleted %s daemon set from %s namespace", name, namespace)

	return nil
}

// DeleteDaemonSetServices deletes services related to daemon set with given name in given namespace.
func DeleteDaemonSetServices(client k8sClient.Interface, namespace, name string) error {
	log.Printf("Deleting services related to %s daemon set from %s namespace", name,
		namespace)

	daemonSet, err := client.Extensions().DaemonSets(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return err
	}

	labelSelector, err := metaV1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
	if err != nil {
		return err
	}

	services, err := GetServicesForDSDeletion(client, labelSelector, namespace)
	if err != nil {
		return err
	}

	for _, service := range services {
		if err := client.Core().Services(namespace).Delete(service.Name, &metaV1.DeleteOptions{}); err != nil {
			return err
		}
	}

	log.Printf("Successfully deleted services related to %s daemon set from %s namespace",
		name, namespace)

	return nil
}
