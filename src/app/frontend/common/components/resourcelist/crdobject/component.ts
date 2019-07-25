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

import {Component, Input} from '@angular/core';
import {HttpParams} from '@angular/common/http';
import {Observable} from 'rxjs';
import {CRDObject, CRDObjectList} from '@api/backendapi';
import {ResourceListBase} from '../../../resources/list';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {ActivatedRoute} from '@angular/router';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';
import {MenuComponent} from '../../list/column/menu/component';

@Component({
  selector: 'kd-crd-object-list',
  templateUrl: './template.html',
})
export class CRDObjectListComponent extends ResourceListBase<CRDObjectList, CRDObject> {
  private readonly endpoint_ = EndpointManager.resource(Resource.crd, true);
  @Input() objects: CRDObject[];
  @Input() initialized: boolean;
  @Input() crdName: string;

  constructor(
    private readonly crdObject_: NamespacedResourceService<CRDObjectList>,
    notifications: NotificationsService,
    private readonly activatedRoute_: ActivatedRoute,
  ) {
    super(`crd/${activatedRoute_.snapshot.params.crdName}`, notifications);
    this.id = ListIdentifier.crdObject;
    this.groupId = ListGroupIdentifier.none;

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);
  }

  getResourceObservable(params?: HttpParams): Observable<CRDObjectList> {
    const {crdName, namespace} = this.activatedRoute_.snapshot.params;
    return this.crdObject_.get(
      this.endpoint_.child(crdName, Resource.crdObject, namespace),
      undefined,
      namespace,
      params,
    );
  }

  map(crdObjectList: CRDObjectList): CRDObject[] {
    return crdObjectList.items;
  }

  getDisplayColumns(): string[] {
    return ['name', 'namespace', 'age'];
  }
}
