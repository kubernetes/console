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

import {HttpParams} from '@angular/common/http';
import {Component, ComponentFactoryResolver, Input} from '@angular/core';
import {Event, Job, JobList} from '@api/backendapi';
import {StateService} from '@uirouter/core';
import {Observable} from 'rxjs/Observable';

import {jobState} from '../../../../resource/workloads/job/state';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NamespaceService} from '../../../services/global/namespace';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {StatusBarColor, StatusBarItem} from '../../statusratiobar/component';
import {ListGroupIdentifiers, ListIdentifiers} from '../groupids';

@Component({
  selector: 'kd-job-list',
  templateUrl: './template.html',
})
export class JobListComponent extends ResourceListWithStatuses<JobList, Job> {
  @Input() title: string;
  @Input() endpoint = EndpointManager.resource(Resource.job, true).list();

  constructor(
      state: StateService, private readonly job_: NamespacedResourceService<JobList>,
      notifications: NotificationsService, resolver: ComponentFactoryResolver,
      private readonly namespaceService_: NamespaceService) {
    super(jobState.name, state, notifications, resolver);
    this.id = ListIdentifiers.job;
    this.groupId = ListGroupIdentifiers.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<JobList> {
    return this.job_.get(this.endpoint, undefined, params);
  }

  map(jobList: JobList): Job[] {
    return jobList.jobs;
  }

  get resourceRatios(): StatusBarItem[] {
    const status = this.resourceList_.status;
    return [
      {
        key: `Running: ${status.running}`,
        color: StatusBarColor.Running,
        value: status.running / this.totalItems * 100,
      },
      {
        key: `Failed: ${status.failed}`,
        color: StatusBarColor.Failed,
        value: status.failed / this.totalItems * 100,
      },
      {
        key: `Pending: ${status.pending}`,
        color: StatusBarColor.Pending,
        value: status.pending / this.totalItems * 100,
      },
      {
        key: `Succeeded: ${status.succeeded}`,
        color: StatusBarColor.Succeeded,
        value: status.succeeded / this.totalItems * 100,
      }
    ];
  }

  isInErrorState(resource: Job): boolean {
    return resource.pods.warnings.length > 0;
  }

  isInPendingState(resource: Job): boolean {
    return resource.pods.warnings.length === 0 && resource.pods.pending > 0;
  }

  isInSuccessState(resource: Job): boolean {
    return resource.pods.warnings.length === 0 && resource.pods.pending === 0;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'age', 'images'];
  }

  hasErrors(job: Job): boolean {
    return job.pods.warnings.length > 0;
  }

  getEvents(job: Job): Event[] {
    return job.pods.warnings;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
