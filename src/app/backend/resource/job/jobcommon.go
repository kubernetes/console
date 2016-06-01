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

package job

import (
	"github.com/kubernetes/dashboard/resource/common"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/batch"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

type JobWithPods struct {
	Job  *batch.Job
	Pods *api.PodList
}

// Returns structure containing Job and Pods for the given job.
func getRawJobWithPods(client client.Interface, namespace, name string) (
	*JobWithPods, error) {
	job, err := client.Extensions().Jobs(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	labelSelector, err := unversioned.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, err
	}

	pods, err := client.Pods(namespace).List(
		api.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: fields.Everything(),
		})

	if err != nil {
		return nil, err
	}

	jobAndPods := &JobWithPods{
		Job:  job,
		Pods: pods,
	}
	return jobAndPods, nil
}

// Retrieves Pod list that belongs to a Job.
func getRawJobPods(client client.Interface, namespace, name string) (*api.PodList, error) {
	jobAndPods, err := getRawJobWithPods(client, namespace, name)
	if err != nil {
		return nil, err
	}
	return jobAndPods.Pods, nil
}

// Returns aggregate information about job pods.
func getJobPodInfo(job *batch.Job, pods []api.Pod) common.PodInfo {
	result := common.PodInfo{}
	for _, pod := range pods {
		switch pod.Status.Phase {
		case api.PodRunning:
			result.Running++
		case api.PodPending:
			result.Pending++
		case api.PodFailed:
			result.Failed++
		}
	}

	return result
}

// Based on given selector returns list of services that are candidates for deletion.
// Services are matched by jobs' label selector. They are deleted if given
// label selector is targeting only 1 job.
func getServicesForDSDeletion(client client.Interface, labelSelector labels.Selector,
	namespace string) ([]api.Service, error) {

	job, err := client.Extensions().Jobs(namespace).List(api.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fields.Everything(),
	})
	if err != nil {
		return nil, err
	}

	// if label selector is targeting only 1 job
	// then we can delete services targeted by this label selector,
	// otherwise we can not delete any services so just return empty list
	if len(job.Items) != 1 {
		return []api.Service{}, nil
	}

	services, err := client.Services(namespace).List(api.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fields.Everything(),
	})
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}
