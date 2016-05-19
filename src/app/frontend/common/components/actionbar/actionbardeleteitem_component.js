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
 * Controller for the action bar delete item component. It adds option to delete item from
 * details page action bar.
 *
 * @final
 */
export class ActionbarDeleteItemController {
  /**
   * @param {!./../../resource/verber_service.VerberService} kdResourceVerberService
   * @param {!./../breadcrumbs/breadcrumbs_service.BreadcrumbsService} kdBreadcrumbsService
   * @param {!ui.router.$state} $state
   * @param {!md.$dialog} $mdDialog
   * @ngInject
   */
  constructor(kdResourceVerberService, kdBreadcrumbsService, $state, $mdDialog) {
    /** @export {string} Initialized from a binding. */
    this.resourceKindName;

    /** @export {!backendApi.TypeMeta} Initialized from a binding. */
    this.typeMeta;

    /** @export {!backendApi.ObjectMeta} Initialized from a binding. */
    this.objectMeta;

    /** @export {boolean} Initialized from a binding. */
    this.isFabButton;

    /** @private {!./../../resource/verber_service.VerberService} */
    this.kdResourceVerberService_ = kdResourceVerberService;

    /** @private {!./../breadcrumbs/breadcrumbs_service.BreadcrumbsService} */
    this.kdBreadcrumbsService_ = kdBreadcrumbsService;

    /** @private {!ui.router.$state}} */
    this.state_ = $state;

    /** @private {!md.$dialog} */
    this.mdDialog_ = $mdDialog;
  }

  /**
   * @export
   */
  remove() {
    this.kdResourceVerberService_
        .showDeleteDialog(this.resourceKindName, this.typeMeta, this.objectMeta)
        .then(
            () => { this.state_.go(this.getFallbackStateName_()); },
            (/** angular.$http.Response|null */ err) => {
              if (err) {
                // Show dialog if there was an error, not user canceling dialog.
                this.mdDialog_.show(this.mdDialog_.alert()
                                        .ok('Ok')
                                        .title(err.statusText || 'Internal server error')
                                        .textContent(err.data || 'Could not delete the resource'));
              }
            });
  }

  /**
   * Returns parent state name based on current state or default state if parent is not found.
   *
   * @return {string}
   * @private
   */
  getFallbackStateName_() {
    return this.kdBreadcrumbsService_.getParentStateName(this.state_['$current']);
  }
}

/**
 * Action bar delete item component should be used only on resource details page in order to
 * add button that allows deletion of this resource.
 *
 * @type {!angular.Component}
 */
export const actionbarDeleteItemComponent = {
  templateUrl: 'common/components/actionbar/actionbardeleteitem.html',
  bindings: {
    'resourceKindName': '@',
    'typeMeta': '<',
    'objectMeta': '<',
    'isFabButton': '<',
  },
  bindToController: true,
  replace: true,
  controller: ActionbarDeleteItemController,
};
