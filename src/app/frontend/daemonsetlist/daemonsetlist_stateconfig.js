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

import {actionbarViewName} from 'chrome/chrome_state';
import {breadcrumbsConfig} from 'common/components/breadcrumbs/breadcrumbs_service';
import {stateName as workloadsState} from 'workloads/workloads_state';
import {stateName as namespaceStateName} from 'common/namespace/namespace_state';

import {DaemonSetListController} from './daemonsetlist_controller';
import {stateName, stateUrl} from './daemonsetlist_state';

/**
 * Configures states for the Daemon Set list view.
 *
 * @param {!ui.router.$stateProvider} $stateProvider
 * @ngInject
 */
export default function stateConfig($stateProvider) {
  $stateProvider.state(stateName, {
    url: stateUrl,
    parent: namespaceStateName,
    resolve: {
      'daemonSetList': resolveDaemonSetList,
    },
    data: {
      [breadcrumbsConfig]: {
        'label': 'Daemon Sets',
        'parent': workloadsState,
      },
    },
    views: {
      '': {
        controller: DaemonSetListController,
        controllerAs: 'ctrl',
        templateUrl: 'daemonsetlist/daemonsetlist.html',
      },
      [actionbarViewName]: {},
    },
  });
}

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.$q.Promise}
 * @ngInject
 */
export function resolveDaemonSetList($resource) {
  /** @type {!angular.Resource<!backendApi.DaemonSetList>} */
  let resource = $resource('api/v1/daemonset');

  return resource.get().$promise;
}
