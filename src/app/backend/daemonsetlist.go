// Copyright 2015 Google Inc. All Rights Reserved.
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

package main

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// DaemonSetList contains a list of Daemon Sets in the cluster.
type DaemonSetList struct {
	// Unordered list of Daemon Sets
	DaemonSets []DaemonSet `json:"daemonSets"`
}

// DaemonSet (aka. Daemon Set) plus zero or more Kubernetes services that
// target the Daemon Set.
type DaemonSet struct {
	// Name of the Daemon Set
	Name string `json:"name"`

	// Namespace this Daemon Set is in.
	Namespace string `json:"namespace"`

	// Human readable description of this Daemon Set.
	Description string `json:"description"`

	// Label of this Daemon Set.
	Labels map[string]string `json:"labels"`

	// Aggregate information about pods belonging to this Daemon Set.
	Pods DaemonSetPodInfo `json:"pods"`

	// Container images of the Daemon Set.
	ContainerImages []string `json:"containerImages"`

	// Time the daemon set was created.
	CreationTime unversioned.Time `json:"creationTime"`

	// Internal endpoints of all Kubernetes services have the same label selector as this Daemon Set.
	InternalEndpoints []Endpoint `json:"internalEndpoints"`

	// External endpoints of all Kubernetes services have the same label selector as this Daemon Set.
	ExternalEndpoints []Endpoint `json:"externalEndpoints"`
}

// GetDaemonSetList returns a list of all Daemon Set in the cluster.
func GetDaemonSetList(client *client.Client, namespace string) (*DaemonSetList, error) {
	log.Printf("Getting list of all daemon sets in the cluster")
	if namespace == "" {
		namespace = api.NamespaceAll
	}

	listEverything := api.ListOptions{
		LabelSelector: labels.Everything(),
		FieldSelector: fields.Everything(),
	}

	daemonSets, err := client.Extensions().DaemonSets(namespace).List(listEverything)

	if err != nil {
		return nil, err
	}

	services, err := client.Services(namespace).List(listEverything)

	if err != nil {
		return nil, err
	}

	pods, err := client.Pods(namespace).List(listEverything)

	if err != nil {
		return nil, err
	}

	// Anonymous callback function to get pods warnings.
	// Function fulfils GetPodsEventWarningsFunc type contract.
	// Based on list of api pods returns list of pod related warning events
	getPodsEventWarningsFn := func(pods []api.Pod) ([]Event, error) {
		errors, err := GetPodsEventWarnings(client, pods)

		if err != nil {
			return nil, err
		}

		return errors, nil
	}

	// Anonymous callback function to get nodes by their names.
	getNodeFn := func(nodeName string) (*api.Node, error) {
		return client.Nodes().Get(nodeName)
	}

	result, err := getDaemonSetList(daemonSets.Items, services.Items,
		pods.Items, getPodsEventWarningsFn, getNodeFn)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Returns a list of all Daemon Set model objects in the cluster, based on all Kubernetes
// Daemon Set and Service API objects.
// The function processes all Daemon Set API objects and finds matching Services for them.
func getDaemonSetList(daemonSets []extensions.DaemonSet,
	services []api.Service, pods []api.Pod, getPodsEventWarningsFn GetPodsEventWarningsFunc,
	getNodeFn GetNodeFunc) (*DaemonSetList, error) {

	daemonSetList := &DaemonSetList{DaemonSets: make([]DaemonSet, 0)}

	for _, daemonSet := range daemonSets {
		var containerImages []string
		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			containerImages = append(containerImages, container.Image)
		}

		matchingServices := getMatchingServicesforDS(services, &daemonSet)
		var internalEndpoints []Endpoint
		var externalEndpoints []Endpoint
		for _, service := range matchingServices {
			internalEndpoints = append(internalEndpoints,
				getInternalEndpoint(service.Name, service.Namespace, service.Spec.Ports))
			externalEndpoints = getExternalEndpointsforDS(daemonSet, pods, service,
				getNodeFn)
		}

		matchingPods := make([]api.Pod, 0)
		for _, pod := range pods {
			if pod.ObjectMeta.Namespace == daemonSet.ObjectMeta.Namespace &&
				isLabelSelectorMatchingforDS(pod.ObjectMeta.Labels, daemonSet.Spec.Selector) {
				matchingPods = append(matchingPods, pod)
			}
		}
		podInfo := getDaemonSetPodInfo(&daemonSet, matchingPods)
		podErrors, err := getPodsEventWarningsFn(matchingPods)

		if err != nil {
			return nil, err
		}

		podInfo.Warnings = podErrors

		daemonSetList.DaemonSets = append(daemonSetList.DaemonSets,
			DaemonSet{
				Name:              daemonSet.ObjectMeta.Name,
				Namespace:         daemonSet.ObjectMeta.Namespace,
				Description:       daemonSet.Annotations[DescriptionAnnotationKey],
				Labels:            daemonSet.ObjectMeta.Labels,
				Pods:              podInfo,
				ContainerImages:   containerImages,
				CreationTime:      daemonSet.ObjectMeta.CreationTimestamp,
				InternalEndpoints: internalEndpoints,
				ExternalEndpoints: externalEndpoints,
			})
	}

	return daemonSetList, nil
}

// Returns all services that target the same Pods (or subset) as the given Daemon Set.
func getMatchingServicesforDS(services []api.Service,
	daemonSet *extensions.DaemonSet) []api.Service {

	var matchingServices []api.Service
	for _, service := range services {
		if service.ObjectMeta.Namespace == daemonSet.ObjectMeta.Namespace &&
			isLabelSelectorMatchingforDS(service.Spec.Selector, daemonSet.Spec.Selector) {

			matchingServices = append(matchingServices, service)
		}
	}
	return matchingServices
}

// Returns true when a Service with the given selector targets the same Pods (or subset) that
// a Daemon Set with the given selector.
func isLabelSelectorMatchingforDS(labelSelector map[string]string,
	testedObjectLabels *unversioned.LabelSelector) bool {

	// If service has no selectors, then assume it targets different Pods.
	if len(labelSelector) == 0 {
		return false
	}
	for label, value := range labelSelector {
		if rsValue, ok := testedObjectLabels.MatchLabels[label]; !ok || rsValue != value {
			return false
		}
	}
	return true
}
