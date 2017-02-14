// Copyright 2016 Google Inc. All Rights Reserved.
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

package horizontalpodautoscalerlist

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/horizontalpodautoscaler"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
)

var (
	apiHpaList = []autoscaling.HorizontalPodAutoscaler{
		{
			ObjectMeta: api.ObjectMeta{Name: "test-hpa1", Namespace: "test-ns"},
			Spec: autoscaling.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscaling.CrossVersionObjectReference{
					Kind: "test-kind1",
					Name: "test-name1",
				},
				MaxReplicas: 3,
			},
			Status: autoscaling.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 1,
				DesiredReplicas: 2,
			},
		}, {
			ObjectMeta: api.ObjectMeta{Name: "test-hpa2", Namespace: "test-ns"},
			Spec: autoscaling.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscaling.CrossVersionObjectReference{
					Kind: "test-kind2",
					Name: "test-name2",
				},
				MaxReplicas: 3,
			},
			Status: autoscaling.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 1,
				DesiredReplicas: 2,
			},
		}, {
			ObjectMeta: api.ObjectMeta{Name: "test-hpa3", Namespace: "test-ns"},
			Spec: autoscaling.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscaling.CrossVersionObjectReference{
					Kind: "test-kind2",
					Name: "test-name2",
				},
				MaxReplicas: 3,
			},
			Status: autoscaling.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 1,
				DesiredReplicas: 2,
			},
		}, {
			ObjectMeta: api.ObjectMeta{Name: "test-hpa4", Namespace: "test-ns"},
			Spec: autoscaling.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscaling.CrossVersionObjectReference{
					Kind: "test-kind2",
					Name: "test-name3",
				},
				MaxReplicas: 3,
			},
			Status: autoscaling.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 1,
				DesiredReplicas: 2,
			},
		},
	}
	ourHpaList = []HorizontalPodAutoscaler{
		{
			ObjectMeta: common.ObjectMeta{Name: "test-hpa1", Namespace: "test-ns"},
			TypeMeta:   common.TypeMeta{Kind: common.ResourceKindHorizontalPodAutoscaler},
			ScaleTargetRef: horizontalpodautoscaler.ScaleTargetRef{
				Kind: "test-kind1",
				Name: "test-name1",
			},
			MaxReplicas: 3,
		}, {
			ObjectMeta: common.ObjectMeta{Name: "test-hpa2", Namespace: "test-ns"},
			TypeMeta:   common.TypeMeta{Kind: common.ResourceKindHorizontalPodAutoscaler},
			ScaleTargetRef: horizontalpodautoscaler.ScaleTargetRef{
				Kind: "test-kind2",
				Name: "test-name2",
			},
			MaxReplicas: 3,
		}, {
			ObjectMeta: common.ObjectMeta{Name: "test-hpa3", Namespace: "test-ns"},
			TypeMeta:   common.TypeMeta{Kind: common.ResourceKindHorizontalPodAutoscaler},
			ScaleTargetRef: horizontalpodautoscaler.ScaleTargetRef{
				Kind: "test-kind2",
				Name: "test-name2",
			},
			MaxReplicas: 3,
		}, {
			ObjectMeta: common.ObjectMeta{Name: "test-hpa4", Namespace: "test-ns"},
			TypeMeta:   common.TypeMeta{Kind: common.ResourceKindHorizontalPodAutoscaler},
			ScaleTargetRef: horizontalpodautoscaler.ScaleTargetRef{
				Kind: "test-kind2",
				Name: "test-name3",
			},
			MaxReplicas: 3,
		},
	}
)

//func GetHorizontalPodAutoscalerList(client k8sClient.Interface, nsQuery *common.NamespaceQuery) (*HorizontalPodAutoscalerList, error) {
func TestGetHorizontalPodAutoscalerList(t *testing.T) {
	cases := []struct {
		expectedActions []string
		hpaList         *autoscaling.HorizontalPodAutoscalerList
		expected        *HorizontalPodAutoscalerList
	}{
		{
			[]string{"list"},
			&autoscaling.HorizontalPodAutoscalerList{
				Items: apiHpaList,
			},
			&HorizontalPodAutoscalerList{
				ListMeta:                 common.ListMeta{TotalItems: 4},
				HorizontalPodAutoscalers: ourHpaList,
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(c.hpaList)

		actual, _ := GetHorizontalPodAutoscalerList(fakeClient, &common.NamespaceQuery{})

		actions := fakeClient.Actions()
		if len(actions) != len(c.expectedActions) {
			t.Errorf("Unexpected actions: %v, expected %d actions got %d", actions,
				len(c.expectedActions), len(actions))
			continue
		}

		for i, verb := range c.expectedActions {
			if actions[i].GetVerb() != verb {
				t.Errorf("Unexpected action: %+v, expected %s",
					actions[i], verb)
			}
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetHorizontalPodAutoscalerList(client, nil) == \ngot: %#v, \nexpected %#v",
				actual, c.expected)
		}
	}
}

//func GetHorizontalPodAutoscalerListForResource(client k8sClient.Interface, namespace, kind, name string) (*HorizontalPodAutoscalerList, error) {
func TestGetHorizontalPodAutoscalerListForResource(t *testing.T) {
	cases := []struct {
		kind, name      string
		expectedActions []string
		hpaList         *autoscaling.HorizontalPodAutoscalerList
		expected        *HorizontalPodAutoscalerList
	}{
		{
			"test-kind1", "test-name1",
			[]string{"list"},
			&autoscaling.HorizontalPodAutoscalerList{
				Items: apiHpaList,
			},
			&HorizontalPodAutoscalerList{
				ListMeta:                 common.ListMeta{TotalItems: 1},
				HorizontalPodAutoscalers: []HorizontalPodAutoscaler{ourHpaList[0]},
			},
		}, {
			"test-kind2", "test-name2",
			[]string{"list"},
			&autoscaling.HorizontalPodAutoscalerList{
				Items: apiHpaList,
			},
			&HorizontalPodAutoscalerList{
				ListMeta:                 common.ListMeta{TotalItems: 2},
				HorizontalPodAutoscalers: []HorizontalPodAutoscaler{ourHpaList[1], ourHpaList[2]},
			},
		}, {
			"test-kind2", "test-name3",
			[]string{"list"},
			&autoscaling.HorizontalPodAutoscalerList{
				Items: apiHpaList,
			},
			&HorizontalPodAutoscalerList{
				ListMeta:                 common.ListMeta{TotalItems: 1},
				HorizontalPodAutoscalers: []HorizontalPodAutoscaler{ourHpaList[3]},
			},
		}, {
			"test-kind1", "test-name2",
			[]string{"list"},
			&autoscaling.HorizontalPodAutoscalerList{
				Items: apiHpaList,
			},
			&HorizontalPodAutoscalerList{
				ListMeta:                 common.ListMeta{TotalItems: 0},
				HorizontalPodAutoscalers: []HorizontalPodAutoscaler{},
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(c.hpaList)

		actual, _ := GetHorizontalPodAutoscalerListForResource(fakeClient, "", c.kind, c.name)

		actions := fakeClient.Actions()
		if len(actions) != len(c.expectedActions) {
			t.Errorf("Unexpected actions: %v, expected %d actions got %d", actions,
				len(c.expectedActions), len(actions))
			continue
		}

		for i, verb := range c.expectedActions {
			if actions[i].GetVerb() != verb {
				t.Errorf("Unexpected action: %+v, expected %s",
					actions[i], verb)
			}
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetHorizontalPodAutoscalerList(client, nil) == \ngot: %#v, \nexpected %#v",
				actual, c.expected)
		}
	}

}
