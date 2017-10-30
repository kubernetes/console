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

/**
 * @final
 */
export default class SettingsEntryController {
  /**
   * @ngInject
   */
  constructor() {
    /** @export {string} Initialized from the scope. */
    this.key;

      /** @export {string} Initialized from the scope. */
      this.desc;
  }

}

/**
 * Labels component definition.
 *
 * @type {!angular.Component}
 */
export const settingsEntryComponent = {
  bindings: {
    'key': '@',
    'desc': '@',
  },
  controller: SettingsEntryController,
  templateUrl: 'settings/settingsentry/settingsentry.html',

};
