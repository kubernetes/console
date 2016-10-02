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

import {LimitRangeListController} from 'limitrangelist/limitrangelist_controller';
import limitRangeListModule from 'limitrangelist/limitrangelist_module';

describe('Limit Range list controller', () => {
  /** @type {!limitrangelist/limitrangelist_controller.LimitRangeListController} */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(limitRangeListModule.name);

    angular.mock.inject(($controller) => {
      ctrl = $controller(LimitRangeListController, {limitRangeList: {items: []}});
    });
  });

  it('should initialize limit range controller', angular.mock.inject(($controller) => {
    let ctrls = {};
    /** @type {!LimitRangeListController} */
    let ctrl = $controller(LimitRangeListController, {limitRangeList: {items: ctrls}});

    expect(ctrl.limitRangeList.items).toBe(ctrls);
  }));

  it('should show zero state', () => { expect(ctrl.shouldShowZeroState()).toBe(true); });

  it('should hide zero state', () => {
    // given
    ctrl.limitRangeList = {items: ['mock']};

    // then
    expect(ctrl.shouldShowZeroState()).toBe(false);
  });
});
