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
	client "k8s.io/client-go/kubernetes"

	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/api/pkg/resource/event"
)

// GetCronJobEvents gets events associated to cron job.
func GetCronJobEvents(client client.Interface, dsQuery *dataselect.DataSelectQuery, namespace, name string) (
	*common.EventList, error) {

	raw, err := event.GetEvents(client, namespace, name)
	if err != nil {
		return event.EmptyEventList, err
	}

	events := event.CreateEventList(raw, dsQuery)
	return &events, nil
}
