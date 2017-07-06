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

package job

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
)

func createJob(name, namespace string, labelSelector map[string]string) *batch.Job {
	var jobCompletions int32
	var parallelism int32

	return &batch.Job{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name, Namespace: namespace, Labels: labelSelector,
		},
		Spec: batch.JobSpec{
			Selector:    &metaV1.LabelSelector{MatchLabels: labelSelector},
			Completions: &jobCompletions,
			Parallelism: &parallelism,
		},
	}
}

func TestGetJobDetail(t *testing.T) {
	var jobCompletions int32
	var parallelism int32

	cases := []struct {
		namespace, name string
		expectedActions []string
		job             *batch.Job
		expected        *JobDetail
	}{
		{
			"ns-1", "job-1",
			[]string{"get", "get", "list", "list", "list", "get", "list", "list"},
			createJob("job-1", "ns-1", map[string]string{"app": "test"}),
			&JobDetail{
				ObjectMeta: api.ObjectMeta{Name: "job-1", Namespace: "ns-1",
					Labels: map[string]string{"app": "test"}},
				TypeMeta: api.TypeMeta{Kind: api.ResourceKindJob},
				PodInfo:  common.PodInfo{Warnings: []common.Event{}},
				PodList: pod.PodList{
					Pods:              []pod.Pod{},
					CumulativeMetrics: make([]metricapi.Metric, 0),
				},
				EventList:   common.EventList{Events: []common.Event{}},
				Parallelism: &jobCompletions,
				Completions: &parallelism,
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(c.job)

		dataselect.DefaultDataSelectWithMetrics.MetricQuery = dataselect.NoMetrics
		actual, _ := GetJobDetail(fakeClient, nil, c.namespace, c.name)

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
			t.Errorf("GetEvents(client,heapsterClient,%#v, %#v) == \ngot: %#v, \nexpected %#v",
				c.namespace, c.name, actual, c.expected)
		}
	}
}

func TestDeleteJob(t *testing.T) {
	var jobCompletions int32
	var parallelism int32

	cases := []struct {
		namespace, name string
		expectedActions []string
		job             *batch.Job
	}{
		{
			"ns-1", "job-1",
			[]string{"get", "list", "delete"},
			createJob("job-1", "ns-1", map[string]string{"app": "test"}),
		},
		{
			namespace: "ns-1", name: "job-1",
			expectedActions: []string{"get", "list", "delete", "delete"},
			job: &batch.Job{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "ns-1", Namespace: "job-1", Labels: map[string]string{"app": "test"},
				},
				Spec: batch.JobSpec{
					Selector:    &metaV1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
					Completions: &jobCompletions,
					Parallelism: &parallelism,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "ns-1",
							Labels:    map[string]string{"app": "test"},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		fakeClient := fake.NewSimpleClientset(c.job)

		DeleteJob(fakeClient, c.namespace, c.name)

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
	}
}
