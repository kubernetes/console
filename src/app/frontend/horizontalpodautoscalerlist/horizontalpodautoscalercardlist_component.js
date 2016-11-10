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

/**
 * @final
 */
export class HorizontalPodAutoscalerCardListController {
  /**
   * @param {!./../common/namespace/namespace_service.NamespaceService} kdNamespaceService
   * @ngInject
   */
  constructor(kdNamespaceService) {
    /** @private {!./../common/namespace/namespace_service.NamespaceService} */
    this.kdNamespaceService_ = kdNamespaceService;
  }

  /**
   * @return {boolean}
   * @export
   */
  areMultipleNamespacesSelected() {
    return this.kdNamespaceService_.areMultipleNamespacesSelected();
  }
}

/**
 * Definition object for the component that displays service card list.
 *
 * @type {!angular.Component}
 */
export const horizontalPodAutoscalerCardListComponent = {
  templateUrl: 'horizontalpodautoscalerlist/horizontalpodautoscalercardlist.html',
  controller: HorizontalPodAutoscalerCardListController,
  bindings: {
    /** {!Array<!backendApi.HorizontalPodAutoscalerList>} */
    'horizontalPodAutoscalerList': '<',
    /** {!angular.Resource} */
    'horizontalPodAutoscalerListResource': '<',
    'showScaleTarget': '<'
  },
};
