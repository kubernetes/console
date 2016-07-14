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

package daemonset

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/event"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// GetDaemonSetEvents gets events associated to pet set.
func GetDaemonSetEvents(client client.Interface, namespace, daemonSetName string) (
	*common.EventList, error) {

	log.Printf("Getting events related to %s pet set in %s namespace", daemonSetName,
		namespace)

	// Get events for pet set.
	rsEvents, err := event.GetEvents(client, namespace, daemonSetName)

	if err != nil {
		return nil, err
	}

	// Get events for pods in pet set.
	podEvents, err := GetDaemonSetPodsEvents(client, namespace, daemonSetName)

	if err != nil {
		return nil, err
	}

	apiEvents := append(rsEvents, podEvents...)

	if !event.IsTypeFilled(apiEvents) {
		apiEvents = event.FillEventsType(apiEvents)
	}

	events := event.ToEventList(apiEvents, namespace)

	log.Printf("Found %d events related to %s daemon set in %s namespace",
		len(events.Events), daemonSetName, namespace)

	return &events, nil
}

// GetDaemonSetPodsEvents gets events associated to pods in pet set.
func GetDaemonSetPodsEvents(client client.Interface, namespace, daemonSetName string) (
	[]api.Event, error) {

	daemonSet, err := client.Extensions().DaemonSets(namespace).Get(daemonSetName)

	if err != nil {
		return nil, err
	}

	podEvents, err := event.GetPodsEvents(client, namespace, daemonSet.Spec.Selector.MatchLabels)

	if err != nil {
		return nil, err
	}

	return podEvents, nil
}
