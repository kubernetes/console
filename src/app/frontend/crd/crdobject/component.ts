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

import {Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {MatButtonToggleGroup} from '@angular/material';
import {HttpClient} from '@angular/common/http';
import {dump as toYaml, load as fromYaml} from 'js-yaml';
import {Subscription} from 'rxjs';
import {CRDObjectDetail} from '@api/backendapi';
import {ActionbarService, ResourceMeta} from '../../common/services/global/actionbar';
import {NamespacedResourceService} from '../../common/services/resource/resource';
import {EndpointManager, Resource} from '../../common/services/resource/endpoint';
import {NotificationsService} from '../../common/services/global/notifications';
import {RawResource} from '../../common/resources/rawresource';

enum Modes {
  JSON = 'json',
  YAML = 'yaml',
}

@Component({selector: 'kd-crd-object-detail', templateUrl: './template.html'})
export class CRDObjectDetailComponent implements OnInit, OnDestroy {
  @ViewChild('group', {static: true}) buttonToggleGroup: MatButtonToggleGroup;

  private objectSubscription_: Subscription;
  private readonly endpoint_ = EndpointManager.resource(Resource.crd, true);
  object: CRDObjectDetail;
  modes = Modes;
  isInitialized = false;
  selectedMode = Modes.YAML;
  objectRaw = '';

  constructor(
    private readonly object_: NamespacedResourceService<CRDObjectDetail>,
    private readonly actionbar_: ActionbarService,
    private readonly activatedRoute_: ActivatedRoute,
    private readonly notifications_: NotificationsService,
    private readonly http_: HttpClient,
  ) {}

  ngOnInit(): void {
    const {crdName, namespace, objectName} = this.activatedRoute_.snapshot.params;
    this.objectSubscription_ = this.object_
      .get(this.endpoint_.child(crdName, objectName, namespace))
      .subscribe((d: CRDObjectDetail) => {
        this.object = d;
        this.notifications_.pushErrors(d.errors);
        this.actionbar_.onInit.emit(new ResourceMeta(d.typeMeta.kind, d.objectMeta, d.typeMeta));
        this.isInitialized = true;

        // Get raw resource
        const url = RawResource.getUrl(this.object.typeMeta, this.object.objectMeta);
        this.http_
          .get(url)
          .toPromise()
          .then(response => {
            this.objectRaw = toYaml(response);
          });
      });

    this.buttonToggleGroup.valueChange.subscribe((selectedMode: Modes) => {
      this.selectedMode = selectedMode;

      if (this.objectRaw) {
        this.updateText();
      }
    });
  }

  ngOnDestroy(): void {
    this.objectSubscription_.unsubscribe();
    this.actionbar_.onDetailsLeave.emit();
  }

  private updateText(): void {
    if (this.selectedMode === Modes.YAML) {
      this.objectRaw = toYaml(JSON.parse(this.objectRaw));
    } else {
      this.objectRaw = this.toRawJSON(fromYaml(this.objectRaw));
    }
  }

  private toRawJSON(object: {}): string {
    return JSON.stringify(object, null, 2);
  }
}
