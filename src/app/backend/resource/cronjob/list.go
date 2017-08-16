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

package cronjob

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	client "k8s.io/client-go/kubernetes"
	batch2 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
)

// CronJobList contains a list of CronJobs in the cluster.
type CronJobList struct {
	ListMeta          api.ListMeta       `json:"listMeta"`
	CumulativeMetrics []metricapi.Metric `json:"cumulativeMetrics"`

	// Unordered list of CronJobs.
	CronJobs []CronJob `json:"cronJobs"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// CronJob is a presentation layer view of Kubernetes CronJob resource. This means it is CronJob plus additional
// augmented data we can get from other sources
type CronJob struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Aggregate information about pods belonging to this CronJob.
	Pods common.PodInfo `json:"pods"`

	// Container images of the CronJob.
	ContainerImages []string `json:"containerImages"`
}

// GetCronJobList returns a list of all CronJobs in the cluster.
func GetCronJobList(client client.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery, metricClient metricapi.MetricClient) (*CronJobList, error) {
	log.Print("Getting list of all cronJobs in the cluster")

	channels := &common.ResourceChannels{
		CronJobList: common.GetCronJobListChannel(client, nsQuery, 1),
		PodList:     common.GetPodListChannel(client, nsQuery, 1),
		EventList:   common.GetEventListChannel(client, nsQuery, 1),
	}

	return GetCronJobListFromChannels(channels, dsQuery, metricClient)
}

// GetCronJobListFromChannels returns a list of all CronJobs in the cluster reading required resource
// list once from the channels.
func GetCronJobListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) (*CronJobList, error) {

	cronJobs := <-channels.CronJobList.List
	err := <-channels.CronJobList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return toCronJobList(cronJobs.Items, nonCriticalErrors, dsQuery, metricClient), nil
}

func toCronJobList(cronJobs []batch2.CronJob, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery,
	metricClient metricapi.MetricClient) *CronJobList {

	cronJobList := &CronJobList{
		CronJobs: make([]CronJob, 0),
		ListMeta: api.ListMeta{TotalItems: len(cronJobs)},
		Errors:   nonCriticalErrors,
	}

	cachedResources := &metricapi.CachedResources{}

	cronJobCells, metricPromises, filteredTotal := dataselect.GenericDataSelectWithFilterAndMetrics(ToCells(cronJobs),
		dsQuery, cachedResources, metricClient)
	cronJobs = FromCells(cronJobCells)
	cronJobList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, cronJob := range cronJobs {
		cronJobList.CronJobs = append(cronJobList.CronJobs, toCronJob(&cronJob))
	}

	cumulativeMetrics, err := metricPromises.GetMetrics()
	cronJobList.CumulativeMetrics = cumulativeMetrics
	if err != nil {
		cronJobList.CumulativeMetrics = make([]metricapi.Metric, 0)
	}

	return cronJobList
}

func toCronJob(cronJob *batch2.CronJob) CronJob {
	return CronJob{
		ObjectMeta: api.NewObjectMeta(cronJob.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindCronJob),
	}
}
