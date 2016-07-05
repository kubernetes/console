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

import {StateParams, stateName as logs} from './../logs_state';

/**
 * Controller for the logs view.
 * @final
 */
export default class LogsToolbarController {
  /**
   * @param {!ui.router.$state} $state
   * @param {!StateParams} $stateParams
   * @param {!../logs_service.LogColorInversionService} logsColorInversionService
   * @ngInject
   */
  constructor($state, $stateParams, podContainers, logsColorInversionService) {
    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /**
     * Service to notify logs controller if any changes on toolbar.
     * @private {!../logs_service.LogColorInversionService}
     */
    this.logsColorInversionService_ = logsColorInversionService;

    /** @export {!Array<string>} */
    this.containers = podContainers.containers;

    /** @export {string} */
    this.container = $stateParams.container || this.containers[0];

    /** @export {../logs_state.StateParams} */
    this.stateParams = $stateParams;

    /** @export */
    this.i18n = i18n;

    /** @export {!Array<string>} */
    this.fontSizes = ['14px', '18px', '24px'];

    /** @export {string} */
    this.fontSize = this.fontSizes[0];
    this.logsColorInversionService_.setFontSize(this.fontSize);
  }

  /**
   * Indicates state of log area color.
   * If false: black text is placed on white area. Otherwise colors are inverted.
   * @export
   * @return {boolean}
   */
  isTextColorInverted() { return this.logsColorInversionService_.getInverted(); }

  /**
   * Execute a code when a user changes the selected option of a container element.
   * @param {string} container
   * @return {string}
   * @export
   */
  onContainerChange(container) {
    return this.state_.go(
        logs,
        new StateParams(this.stateParams.objectNamespace, this.stateParams.objectName, container));
  }

  /**
   * Execute a code when a user changes the selected option for console font size.
   * @export
   */
  onFontSizeChange() { this.logsColorInversionService_.setFontSize(this.fontSize); }

  /**
   * Return proper style class for icon.
   * @export
   * @returns {string}
   */
  getStyleClass() {
    const logsTextColor = 'kd-logs-color-icon';
    if (this.isTextColorInverted()) {
      return `${logsTextColor}-invert`;
    }
    return `${logsTextColor}`;
  }

  /**
   * Execute a code when a user changes the selected option for console color.
   * @export
   */
  onTextColorChange() { this.logsColorInversionService_.invert(); }

  /**
   * Find Pod by name.
   * Return object or undefined if can not find a object.
   * @param {!Array<!backendApi.ReplicationControllerPodWithContainers>} array
   * @param {!string} name
   * @return {!backendApi.ReplicationControllerPodWithContainers|undefined}
   * @private
   */
  findPodByName_(array, name) {
    for (let i = 0; i < array.length; i++) {
      if (array[i].name === name) {
        return array[i];
      }
    }
    return undefined;
  }

  /**
   * Find Container by name.
   * Return object or undefined if can not find a object.
   * @param {!Array<!backendApi.PodContainer>} array
   * @param {string} name
   * @return {!backendApi.PodContainer}
   * @private
   */
  initializeContainer_(array, name) {
    let container = undefined;
    for (let i = 0; i < array.length; i++) {
      if (array[i].name === name) {
        container = array[i];
        break;
      }
    }
    if (!container) {
      container = array[0];
    }
    return container;
  }
}

const i18n = {
  /** @export {string} @desc Label 'Pod' on the toolbar of the logs page. Ends with colon. */
  MSG_LOGS_POD_LABEL: goog.getMsg('Pod:'),
  /** @export {string} @desc Label 'Container' on the toolbar of the logs page. Ends with colon. */
  MSG_LOGS_CONTAINER_LABEL: goog.getMsg('Container:'),
  /** @export {string} @desc Label 'Font size' on the toolbar of the logs page. Ends with colon. */
  MSG_LOGS_FONT_SIZE_LABEL: goog.getMsg('Font size:'),
};
