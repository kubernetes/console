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

import {ZeroStateController} from 'common/components/zerostate/zerostate_component';

/**
 * @final
 */
export class ServiceListController {
  /**
   * @param {!backendApi.ServiceList} serviceList
   * @param {!ui.router.$state} $state
   * @ngInject
   */
  constructor(serviceList, $state) {
    /** @export {!backendApi.ServiceList} */
    this.serviceList = serviceList;

    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /** Initializes state 'showZeroState' custom data based on given resources array */
    ZeroStateController.showZeroState(this.state_, this.serviceList.services);
  }
}
