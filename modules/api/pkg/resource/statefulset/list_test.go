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

package statefulset

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

func getReplicasPointer(replicas int32) *int32 {
	return &replicas
}

func TestGetStatefulSetListFromChannels(t *testing.T) {
	controller := true
	cases := []struct {
		k8sRs         apps.StatefulSetList
		k8sRsError    error
		pods          *v1.PodList
		expected      *StatefulSetList
		expectedError error
	}{
		{
			apps.StatefulSetList{},
			nil,
			&v1.PodList{},
			&StatefulSetList{
				ListMeta:          types.ListMeta{},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				StatefulSets:      []StatefulSet{},
				Errors:            []error{},
			},
			nil,
		},
		{
			apps.StatefulSetList{},
			errors.NewInvalid("MyCustomError"),
			&v1.PodList{},
			nil,
			errors.NewInvalid("MyCustomError"),
		},
		{
			apps.StatefulSetList{},
			&k8serrors.StatusError{},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{},
		},
		{
			apps.StatefulSetList{},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
		},
		{
			apps.StatefulSetList{},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
		},
		{
			apps.StatefulSetList{
				Items: []apps.StatefulSet{{
					ObjectMeta: metaV1.ObjectMeta{
						Name:              "rs-name",
						Namespace:         "rs-namespace",
						UID:               "uid",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					Spec: apps.StatefulSetSpec{
						Selector: &metaV1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
						Replicas: getReplicasPointer(21),
					},
					Status: apps.StatefulSetStatus{
						Replicas: 7,
					},
				}},
			},
			nil,
			&v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metaV1.ObjectMeta{
							Namespace: "rs-namespace",
							Labels:    map[string]string{"foo": "baz"},
							OwnerReferences: []metaV1.OwnerReference{
								{
									Name:       "rs-name-wrong",
									UID:        "uid-wrong",
									Controller: &controller,
								},
							},
						},
						Status: v1.PodStatus{Phase: v1.PodFailed},
					},
					{
						ObjectMeta: metaV1.ObjectMeta{
							Namespace: "rs-namespace",
							Labels:    map[string]string{"foo": "bar"},
							OwnerReferences: []metaV1.OwnerReference{
								{
									Name:       "rs-name",
									UID:        "uid",
									Controller: &controller,
								},
							},
						},
						Status: v1.PodStatus{Phase: v1.PodFailed},
					},
				},
			},
			&StatefulSetList{
				ListMeta:          types.ListMeta{TotalItems: 1},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				Status:            common.ResourceStatus{Running: 1},
				StatefulSets: []StatefulSet{{
					ObjectMeta: types.ObjectMeta{
						Name:              "rs-name",
						Namespace:         "rs-namespace",
						UID:               "uid",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					TypeMeta: types.TypeMeta{
						Kind:        types.ResourceKindStatefulSet,
						Scalable:    true,
						Restartable: true,
					},
					Pods: common.PodInfo{
						Current:  7,
						Desired:  getReplicasPointer(21),
						Failed:   1,
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
			StatefulSetList: common.StatefulSetListChannel{
				List:  make(chan *apps.StatefulSetList, 1),
				Error: make(chan error, 1),
			},
			NodeList: common.NodeListChannel{
				List:  make(chan *v1.NodeList, 1),
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

		channels.StatefulSetList.Error <- c.k8sRsError
		channels.StatefulSetList.List <- &c.k8sRs

		channels.NodeList.List <- &v1.NodeList{}
		channels.NodeList.Error <- nil

		channels.ServiceList.List <- &v1.ServiceList{}
		channels.ServiceList.Error <- nil

		channels.PodList.List <- c.pods
		channels.PodList.Error <- nil

		channels.EventList.List <- &v1.EventList{}
		channels.EventList.Error <- nil

		actual, err := GetStatefulSetListFromChannels(channels, dataselect.NoDataSelect, nil)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetStatefulSetListChannels() ==\n          %#v\nExpected: %#v", actual, c.expected)
		}
		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("GetStatefulSetListChannels() ==\n          %#v\nExpected: %#v", err, c.expectedError)
		}
	}
}
