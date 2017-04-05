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
export class ActionBarController {
  /**
   * @param {!backendApi.DeploymentDetail} deploymentDetail
   * @param {!common/scaling/scaleresource_service.ScaleService} kdScaleService
   * @ngInject
   */
  constructor(deploymentDetail, kdScaleService) {
    /** @export {!backendApi.DeploymentDetail} */
    this.details = deploymentDetail;

    /** @private {!common/scaling/scaleresource_service.ScaleService} */
    this.kdScaleService_ = kdScaleService;
  }

  /**
   * Handles update of replicas count in deployment dialog.
   * @export
   */
  handleScaleResourceDialog() {
    this.kdScaleService_.showScaleDialog(
        this.details.objectMeta.namespace, this.details.objectMeta.name,
        this.details.statusInfo.replicas, this.details.statusInfo.available,
        this.details.typeMeta.kind);
  }
}
