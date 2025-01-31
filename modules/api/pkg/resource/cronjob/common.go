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

package cronjob

import (
	batch "k8s.io/api/batch/v1"

	metricapi "k8s.io/dashboard/api/pkg/integration/metric/api"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/types"
)

// The code below allows to perform complex data section on []batch.CronJob

type CronJobCell batch.CronJob

func (self CronJobCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
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

func (self CronJobCell) GetResourceSelector() *metricapi.ResourceSelector {
	return &metricapi.ResourceSelector{
		Namespace:    self.ObjectMeta.Namespace,
		ResourceType: types.ResourceKindCronJob,
		ResourceName: self.ObjectMeta.Name,
		UID:          self.UID,
	}
}

func ToCells(std []batch.CronJob) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = CronJobCell(std[i])
	}
	return cells
}

func FromCells(cells []dataselect.DataCell) []batch.CronJob {
	std := make([]batch.CronJob, len(cells))
	for i := range std {
		std[i] = batch.CronJob(cells[i].(CronJobCell))
	}
	return std
}

func getStatus(list *batch.CronJobList) common.ResourceStatus {
	info := common.ResourceStatus{}
	if list == nil {
		return info
	}

	for _, cronJob := range list.Items {
		if cronJob.Spec.Suspend != nil && !(*cronJob.Spec.Suspend) {
			info.Running++
		} else {
			info.Failed++
		}
	}

	return info
}

func getContainerImages(cronJob *batch.CronJob) []string {
	podSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec
	result := make([]string, 0)

	for _, container := range podSpec.Containers {
		result = append(result, container.Image)
	}

	return result
}
