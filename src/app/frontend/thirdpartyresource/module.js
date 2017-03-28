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

import chromeModule from 'chrome/module';
import componentsModule from 'common/components/components_module';
import filtersModule from 'common/filters/filters_module';

import {tprInfoComponent} from './detail/info_component';
import {objectCardComponent} from './detail/objectcard_component';
import {objectListComponent} from './detail/objectlist_component';
import {thirdPartyResourceObjectsResource} from './detail/stateconfig';
import {thirdPartyResourceCardComponent} from './list/card_component';
import {thirdPartyResourceCardListComponent} from './list/cardlist_component';
import {thirdPartyResourceListResource} from './list/stateconfig';
import stateConfig from './stateconfig';

/**
 * Angular module for the Storage Class resource.
 */
export default angular
    .module(
        'kubernetesDashboard.thirdPartyResource',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          chromeModule.name,
          componentsModule.name,
          filtersModule.name,
        ])
    .config(stateConfig)
    .component('kdObjectCard', objectCardComponent)
    .component('kdThirdPartyResourceCard', thirdPartyResourceCardComponent)
    .component('kdThirdPartyResourceCardList', thirdPartyResourceCardListComponent)
    .component('kdThirdPartyResourceInfo', tprInfoComponent)
    .component('kdThirdPartyResourceObjects', objectListComponent)
    .factory('kdThirdPartyResourceListResource', thirdPartyResourceListResource)
    .factory('kdThirdPartyResourceObjectsResource', thirdPartyResourceObjectsResource);
