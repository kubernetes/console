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

import {actionbarViewName, stateName as chromeStateName} from 'chrome/chrome_state';
import {breadcrumbsConfig} from 'common/components/breadcrumbs/breadcrumbs_service';
import {PaginationService} from 'common/pagination/pagination_service';

import {NodeListController} from './nodelist_controller';
import {stateName, stateUrl} from './nodelist_state';

/**
 * Configures states for the service view.
 *
 * @param {!ui.router.$stateProvider} $stateProvider
 * @ngInject
 */
export default function stateConfig($stateProvider) {
  $stateProvider.state(stateName, {
    url: stateUrl,
    parent: chromeStateName,
    resolve: {
      'nodeList': resolveNodeList,
    },
    data: {
      [breadcrumbsConfig]: {
        'label': i18n.MSG_BREADCRUMBS_NODES_LABEL,
      },
    },
    views: {
      '': {
        controller: NodeListController,
        controllerAs: '$ctrl',
        templateUrl: 'nodelist/nodelist.html',
      },
      [actionbarViewName]: {},
    },
  });
}

/**
 * @param {!angular.Resource} kdNodeListResource
 * @param {!./../chrome/chrome_state.StateParams} $stateParams
 * @return {!angular.$q.Promise}
 * @ngInject
 */
export function resolveNodeList(kdNodeListResource, $stateParams) {
  /** @type {!backendApi.PaginationQuery} */
  let query = PaginationService.getDefaultResourceQuery($stateParams.namespace);
  return kdNodeListResource.get(query).$promise;
}

const i18n = {
  /** @type {string} @desc Label 'Nodes' that appears as a breadcrumbs on the action bar. */
  MSG_BREADCRUMBS_NODES_LABEL: goog.getMsg('Nodes'),
};
