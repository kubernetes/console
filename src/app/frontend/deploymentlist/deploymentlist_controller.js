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
 * Controller for the replica set list view.
 *
 * @final
 */
export class DeploymentListController {
  /**
   * @param {!backendApi.DeploymentList} deploymentList
   * @param {!angular.Resource} kdDeploymentListResource
   * @ngInject
   */
  constructor(deploymentList, kdDeploymentListResource) {
    /** @export {!backendApi.DeploymentList} */
    this.deploymentList = deploymentList;

    /** @export {!angular.Resource} */
    this.deploymentListResource = kdDeploymentListResource;

    /** @export */
    this.i18n = i18n;
  }

  /**
   * @return {boolean}
   * @export
   */
  shouldShowZeroState() { return this.deploymentList.deployments.length === 0; }
}

const i18n = {
  /** @export {string} @desc Title for graph card displaying CPU metric of deployments. */
  MSG_DEPLOYMENT_LIST_CPU_GRAPH_CARD_TITLE: goog.getMsg('CPU usage history'),
  /** @export {string} @desc Title for graph card displaying memory metric of deployments. */
  MSG_DEPLOYMENT_LIST_MEMORY_GRAPH_CARD_TITLE: goog.getMsg('Memory usage history'),
};
