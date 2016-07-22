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

import daemonsetListModule from 'daemonsetlist/daemonsetlist_module';

describe('Daemon Set Card List controller', () => {
  /**
   * @type {!daemonsetlist/daemonsetcardlist_component.DaemonSetCardListController}
   */
  let ctrl;
  /**
   * @type {!./../common/namespace/namespace_service.NamespaceService}
   */
  let data;

  beforeEach(() => {
    angular.mock.module(daemonsetListModule.name);

    angular.mock.inject(($componentController, kdNamespaceService) => {
      /** @type {!./../common/namespace/namespace_service.NamespaceService} */
      data = kdNamespaceService;
      /** @type {!DaemonSetCardListController} */
      ctrl = $componentController('kdDaemonSetCardList', {kdNamespaceService_: data});
    });
  });

  it('should instantiate the controller properly', () => { expect(ctrl).not.toBeUndefined(); });

  it('should return the value from Namespace service', () => {
    expect(ctrl.areMultipleNamespacesSelected()).toBe(data.areMultipleNamespacesSelected());
  });
});
