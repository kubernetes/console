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

import {Component, OnInit} from '@angular/core';
import {StateService} from '@uirouter/core';
import {Subscription} from 'rxjs/Subscription';

import {ExecStateParams} from '../../../../../common/params/params';
import {ActionbarService, ResourceMeta} from '../../../../../common/services/global/actionbar';
import {KdStateService} from '../../../../../common/services/global/state';

@Component({templateUrl: './template.html'})
export class ActionbarComponent implements OnInit {
  isInitialized = false;
  resourceMeta: ResourceMeta;
  resourceMetaSubscription_: Subscription;

  constructor(
      private readonly actionbar_: ActionbarService, private readonly kdState_: KdStateService) {}

  ngOnInit(): void {
    this.resourceMetaSubscription_ =
        this.actionbar_.onInit.subscribe((resourceMeta: ResourceMeta) => {
          this.resourceMeta = resourceMeta;
          this.isInitialized = true;
        });
  }

  ngOnDestroy(): void {
    this.resourceMetaSubscription_.unsubscribe();
  }

  viewLogs(): void {
    // TODO
  }

  exec(): void {
    const shellLink = this.kdState_.href(
        'shell', this.resourceMeta.objectMeta.name, this.resourceMeta.objectMeta.namespace);
    window.open(shellLink);
  }
}
