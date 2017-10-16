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

import {namespaceParam} from '../../chrome/state';
import {stateName as overview} from '../../overview/state';

/** @final */
class NamespaceChangeInfoDialogController {
  /**
   * @param {!md.$dialog} $mdDialog
   * @param {string} newNamespace
   * @param {!./../state/service.FutureStateService} kdFutureStateService
   * @param {!ui.router.$state} $state
   * @param {!../history/service.HistoryService} kdHistoryService
   * @ngInject
   */
  constructor($mdDialog, newNamespace, kdFutureStateService, $state, kdHistoryService) {
    /** @private {!md.$dialog} */
    this.mdDialog_ = $mdDialog;

    /** @private {!./../state/service.FutureStateService}} */
    this.futureStateService_ = kdFutureStateService;

    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /** @private {string} */
    this.newNamespace_ = newNamespace;

    /** @private {!../history/service.HistoryService} */
    this.kdHistoryService_ = kdHistoryService;
  }

  /**
   * Cancels namespace change and redirects to last state.
   *
   * @export
   */
  cancel() {
    this.mdDialog_.cancel();
    this.kdHistoryService_.back(overview);
  }

  /** @export */
  changeNamespace() {
    this.mdDialog_.hide();
    this.state_.go(this.futureStateService_.state, {[namespaceParam]: this.newNamespace_});
  }
}

/**
 * Displays new namespace change info dialog.
 *
 * @param {!md.$dialog} mdDialog
 * @param {string} newNamespace
 * @return {!angular.$q.Promise}
 */
export function showNamespaceChangeInfoDialog(mdDialog, newNamespace) {
  return mdDialog.show({
    controller: NamespaceChangeInfoDialogController,
    controllerAs: '$ctrl',
    clickOutsideToClose: false,
    templateUrl: 'common/namespace/dialog.html',
    locals: {
      'newNamespace': newNamespace,
    },
  });
}
