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
	"reflect"
	"testing"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metricapi "k8s.io/dashboard/api/pkg/integration/metric/api"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/errors"
	"k8s.io/dashboard/types"
)

func TestGetDaemonSetListFromChannels(t *testing.T) {
	cases := []struct {
		k8sDaemonSet      apps.DaemonSetList
		k8sDaemonSetError error
		pods              *v1.PodList
		expected          *DaemonSetList
		expectedError     error
	}{
		{
			apps.DaemonSetList{},
			nil,
			&v1.PodList{},
			&DaemonSetList{
				ListMeta:          types.ListMeta{},
				DaemonSets:        []DaemonSet{},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				Errors:            []error{},
			},
			nil,
		},
		{
			apps.DaemonSetList{},
			errors.NewInvalid("MyCustomError"),
			&v1.PodList{},
			nil,
			errors.NewInvalid("MyCustomError"),
		},
		{
			apps.DaemonSetList{},
			&k8serrors.StatusError{},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{},
		},
		{
			apps.DaemonSetList{},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
		},
		{
			apps.DaemonSetList{Items: []apps.DaemonSet{}},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
		},
		{
			apps.DaemonSetList{
				Items: []apps.DaemonSet{{
					ObjectMeta: metaV1.ObjectMeta{
						Name:              "ds-name",
						Namespace:         "ds-namespace",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					Spec: apps.DaemonSetSpec{
						Selector: &metaV1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
					},
					Status: apps.DaemonSetStatus{DesiredNumberScheduled: 7},
				}},
			},
			nil,
			&v1.PodList{},
			&DaemonSetList{
				ListMeta:          types.ListMeta{TotalItems: 1},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				Status:            common.ResourceStatus{Running: 1},
				DaemonSets: []DaemonSet{{
					ObjectMeta: types.ObjectMeta{
						Name:              "ds-name",
						Namespace:         "ds-namespace",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					TypeMeta: types.TypeMeta{Kind: types.ResourceKindDaemonSet, Restartable: true},
					Pods: common.PodInfo{
						Current:  0,
						Failed:   0,
						Warnings: []common.Event{},
					},
				}},
				Errors: []error{},
			},
			nil,
		},
	}

	for _, c := range cases {
		channels := &common.ResourceChannels{
			DaemonSetList: common.DaemonSetListChannel{
				List:  make(chan *apps.DaemonSetList, 1),
				Error: make(chan error, 1),
			},
			ServiceList: common.ServiceListChannel{
				List:  make(chan *v1.ServiceList, 1),
				Error: make(chan error, 1),
			},
			PodList: common.PodListChannel{
				List:  make(chan *v1.PodList, 1),
				Error: make(chan error, 1),
			},
			EventList: common.EventListChannel{
				List:  make(chan *v1.EventList, 1),
				Error: make(chan error, 1),
			},
		}

		channels.DaemonSetList.Error <- c.k8sDaemonSetError
		channels.DaemonSetList.List <- &c.k8sDaemonSet

		channels.ServiceList.List <- &v1.ServiceList{}
		channels.ServiceList.Error <- nil

		channels.PodList.List <- c.pods
		channels.PodList.Error <- nil

		channels.EventList.List <- &v1.EventList{}
		channels.EventList.Error <- nil

		actual, err := GetDaemonSetListFromChannels(channels, dataselect.NoDataSelect, nil)

		// Rewrite address of desired number of pods.
		if actual != nil {
			for i := range actual.DaemonSets {
				c.expected.DaemonSets[i].Pods.Desired = actual.DaemonSets[i].Pods.Desired
			}
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetDaemonSetListFromChannels() ==\n          %#v\nExpected: %#v", actual, c.expected)
		}
		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("GetDaemonSetListFromChannels() ==\n          %#v\nExpected: %#v", err, c.expectedError)
		}
	}
}

func TestToDaemonSetList(t *testing.T) {
	events := []v1.Event{}
	controller := true
	var desired int32 = 1
	var desire2 int32 = 2
	var desire3 int32 = 3
	validPodMeta := metaV1.ObjectMeta{
		Namespace: "namespace-1",
		OwnerReferences: []metaV1.OwnerReference{
			{
				Kind:       "DaemonSet",
				Name:       "my-name-1",
				UID:        "uid-1",
				Controller: &controller,
			},
		},
	}
	diffPodMeta := metaV1.ObjectMeta{
		Namespace: "namespace-2",
		OwnerReferences: []metaV1.OwnerReference{
			{
				Kind:       "DaemonSet",
				Name:       "my-name-2",
				UID:        "uid-2",
				Controller: &controller,
			},
		},
	}

	cases := []struct {
		daemonSets []apps.DaemonSet
		services   []v1.Service
		pods       []v1.Pod
		nodes      []v1.Node
		expected   *DaemonSetList
	}{
		{nil,
			nil,
			nil,
			nil,
			&DaemonSetList{
				DaemonSets:        []DaemonSet{},
				CumulativeMetrics: make([]metricapi.Metric, 0)},
		}, {
			[]apps.DaemonSet{
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-1",
						Namespace: "namespace-1",
						UID:       "uid-1",
					},
					Spec: apps.DaemonSetSpec{
						Selector: &metaV1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-name-1"},
						},
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{Containers: []v1.Container{{Image: "my-container-image-1"}},
								InitContainers: []v1.Container{{Image: "my-init-container-image-1"}}},
						},
					},
					Status: apps.DaemonSetStatus{
						DesiredNumberScheduled: desire3,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-2",
						Namespace: "namespace-2",
					},
					Spec: apps.DaemonSetSpec{
						Selector: &metaV1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-name-2", "ver": "2"},
						},
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{Containers: []v1.Container{{Image: "my-container-image-2"}},
								InitContainers: []v1.Container{{Image: "my-init-container-image-2"}}},
						},
					},
					Status: apps.DaemonSetStatus{
						DesiredNumberScheduled: desired,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-3",
						Namespace: "namespace-3",
					},
					Spec: apps.DaemonSetSpec{
						Selector: &metaV1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-name-3", "ver": "3"},
						},
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{Containers: []v1.Container{{Image: "my-container-image-3"}},
								InitContainers: []v1.Container{{Image: "my-init-container-image-3"}}},
						},
					},
					Status: apps.DaemonSetStatus{
						DesiredNumberScheduled: desire2,
					},
				},
			},
			[]v1.Service{
				{
					Spec: v1.ServiceSpec{Selector: map[string]string{"app": "my-name-1"}},
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-1",
						Namespace: "namespace-1",
					},
				},
				{
					Spec: v1.ServiceSpec{Selector: map[string]string{"app": "my-name-2", "ver": "2"}},
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-2",
						Namespace: "namespace-2",
					},
				},
			},
			[]v1.Pod{
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				{
					ObjectMeta: diffPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				},
				{
					ObjectMeta: validPodMeta,
					Status: v1.PodStatus{
						Phase: v1.PodUnknown,
					},
				},
			},
			[]v1.Node{{
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeExternalIP,
							Address: "192.168.1.108",
						},
					},
				},
			},
			},
			&DaemonSetList{
				ListMeta:          types.ListMeta{TotalItems: 3},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				DaemonSets: []DaemonSet{
					{
						ObjectMeta: types.ObjectMeta{
							Name:      "my-app-1",
							Namespace: "namespace-1",
							UID:       "uid-1",
						},
						TypeMeta:            types.TypeMeta{Kind: types.ResourceKindDaemonSet, Restartable: true},
						ContainerImages:     []string{"my-container-image-1"},
						InitContainerImages: []string{"my-init-container-image-1"},
						Pods: common.PodInfo{
							Desired:   &desire3,
							Failed:    2,
							Pending:   1,
							Running:   1,
							Succeeded: 1,
							Warnings:  []common.Event{},
						},
					}, {
						ObjectMeta: types.ObjectMeta{
							Name:      "my-app-2",
							Namespace: "namespace-2",
						},
						TypeMeta:            types.TypeMeta{Kind: types.ResourceKindDaemonSet, Restartable: true},
						ContainerImages:     []string{"my-container-image-2"},
						InitContainerImages: []string{"my-init-container-image-2"},
						Pods: common.PodInfo{
							Desired:  &desired,
							Warnings: []common.Event{},
						},
					}, {
						ObjectMeta: types.ObjectMeta{
							Name:      "my-app-3",
							Namespace: "namespace-3",
						},
						TypeMeta:            types.TypeMeta{Kind: types.ResourceKindDaemonSet, Restartable: true},
						ContainerImages:     []string{"my-container-image-3"},
						InitContainerImages: []string{"my-init-container-image-3"},
						Pods: common.PodInfo{
							Desired:  &desire2,
							Warnings: []common.Event{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		actual := toDaemonSetList(c.daemonSets, c.pods, events, nil, dataselect.NoDataSelect, nil)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("toDaemonSetList(%#v, %#v, %#v) == \n%#v\nexpected \n%#v\n", c.daemonSets, c.services, events, actual, c.expected)
		}
	}
}
