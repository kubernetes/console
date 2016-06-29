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

import {ReplicationControllerListController} from 'replicationcontrollerlist/replicationcontrollerlist_controller';
import replicationControllerListModule from 'replicationcontrollerlist/replicationcontrollerlist_module';

describe('Replication controller list controller', () => {
  /** @type
   * {!replicationcontrollerlist/replicationcontrollerlist_controller.ReplicationControllerListController}
   */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(replicationControllerListModule.name);

    angular.mock.inject(($controller) => {
      ctrl = $controller(
          ReplicationControllerListController,
          {replicationControllerList: {replicationControllers: []}});
    });
  });

  it('should initialize replication controller list', angular.mock.inject(($controller) => {
    let ctrls = {};
    /** @type {!ReplicationControllerListController} */
    let ctrl = $controller(
        ReplicationControllerListController,
        {replicationControllerList: {replicationControllers: ctrls}});

    expect(ctrl.replicationControllerList.replicationControllers).toBe(ctrls);
  }));

  it('should show zero state', () => { expect(ctrl.shouldShowZeroState()).toBeTruthy(); });

  it('should hide zero state', () => {
    // given
    ctrl.replicationControllerList = {replicationControllers: ['mock']};

    // then
    expect(ctrl.shouldShowZeroState()).toBeFalsy();
  });
});
