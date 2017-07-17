// Copyright 2017 The Kubernetes Dashboard Authors.
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

export default class WarningsController {
  /**
   * Constructs warnings controller.
   * @ngInject
   */
  constructor() {
    /**
     * Limit of displayed warnings used by default.
     * @type {number}
     */
    this.defaultLimit = 3;

    /**
     * Currently used limit of displayed warnings.
     * @type {number}
     */
    this.limit = this.defaultLimit;
  }

  /**
   * Toggles between two states of warnings display: show all and show only specific number of
   * warnings.
   *
   * @export
   */
  toggleWarnings() {
    if (this.limit) {
      this.limit = undefined;
    } else {
      this.limit = this.defaultLimit;
    }
  }

  /**
   * Dismisses single warning.
   *
   * @param {number} index
   * @export
   */
  dismissWarning(index) {
    this.warnings.splice(index, 1);
  }

  /**
   * Dismisses all warnings at once.
   *
   * @export
   */
  dismissWarnings() {
    this.warnings = [];
  }

  /**
   * @export
   * @return {string}
   */
  getShowMoreLabel() {
    /** @type {string} @desc Show more warnings button label. */
    let MSG_MORE_WARNINGS_LABEL =
        goog.getMsg('Show {$count} more', {'count': this.warnings.length - this.defaultLimit});
    return MSG_MORE_WARNINGS_LABEL;
  }

  /**
   * @export
   * @return {string}
   */
  getShowLessLabel() {
    /** @type {string} @desc Show less warnings button label. */
    let MSG_LESS_WARNINGS_LABEL =
        goog.getMsg('Show {$count} less', {'count': this.warnings.length - this.defaultLimit});
    return MSG_LESS_WARNINGS_LABEL;
  }
}

/**
 * @type {!angular.Component}
 */
export const warningsComponent = {
  bindings: {
    'warnings': '<',
  },
  controller: WarningsController,
  templateUrl: 'common/components/warnings/warnings.html',
};
