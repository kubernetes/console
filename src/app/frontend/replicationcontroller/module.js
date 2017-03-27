// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import chromeModule from 'chrome/chrome_module';
import componentsModule from 'common/components/components_module';
import csrfTokenModule from 'common/csrftoken/csrftoken_module';
import filtersModule from 'common/filters/filters_module';
import namespaceModule from 'common/namespace/namespace_module';
import eventsModule from 'events/events_module';
import podModule from 'pod/module';

import {ReplicationControllerService} from './detail/delete_service';
import {replicationControllerInfoComponent} from './detail/info_component';
import {replicationControllerEventsResource, replicationControllerPodsResource} from './detail/stateconfig';
import {replicationControllerResource, replicationControllerServicesResource} from './detail/stateconfig';
import {replicationControllerCardComponent} from './list/card_component';
import {replicationControllerCardListComponent} from './list/cardlist_component';
import {replicationControllerCardMenuComponent} from './list/cardmenu_component';
import {replicationControllerListResource} from './list/stateconfig';
import stateConfig from './stateconfig';

/**
 * Angular module for the Deployment resource.
 */
export default angular
    .module(
        'kubernetesDashboard.replicationController',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          chromeModule.name,
          componentsModule.name,
          eventsModule.name,
          filtersModule.name,
          namespaceModule.name,
          csrfTokenModule.name,
          podModule.name,
        ])
    .config(stateConfig)
    .component('kdReplicationControllerCard', replicationControllerCardComponent)
    .component('kdReplicationControllerCardList', replicationControllerCardListComponent)
    .component('kdReplicationControllerCardMenu', replicationControllerCardMenuComponent)
    .component('kdReplicationControllerInfo', replicationControllerInfoComponent)
    .service('kdReplicationControllerService', ReplicationControllerService)
    .factory('kdRCListResource', replicationControllerListResource)
    .factory('kdRCResource', replicationControllerResource)
    .factory('kdRCPodsResource', replicationControllerPodsResource)
    .factory('kdRCEventsResource', replicationControllerEventsResource)
    .factory('kdRCServicesResource', replicationControllerServicesResource);
