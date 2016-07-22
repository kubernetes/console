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

import deploymentListModule from 'deploymentlist/deploymentlist_module';

describe('Deployment Card List controller', () => {
  /**
   * @type {!deploymentlist/deploymentcardlist_component.DeploymentCardListController}
   */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(deploymentListModule.name);

    angular.mock.inject(($componentController, kdNamespaceService) => {
      /** @type {!./../common/namespace/namespace_service.NamespaceService} */
      let data = kdNamespaceService;
      /** @type {!DeploymentCardListController} */
      ctrl = $componentController('kdDeploymentCardList', {kdNamespaceService_: data});
      expect(ctrl.isMultipleNamespaces).toBe(data.getMultipleNamespacesSelected());
    });
  });

  it('should instantiate the controller properly', () => { expect(ctrl).not.toBeUndefined(); });
});
