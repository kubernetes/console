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

import chromeModule from 'chrome/chrome_module';
import stateConfig from './secretlist_stateconfig';
import paginationModule from 'common/pagination/pagination_module';
import componentsModule from 'common/components/components_module';
import filtersModule from 'common/filters/filters_module';
import {secretCardListComponent} from './secretcardlist_component';
import {secretCardComponent} from './secretcard_component';

/**
 * Angular module for the Secrets list view.
 *
 * The view shows Secrets running in the cluster and allows to manage them.
 */
export default angular
    .module(
        'kubernetesDashboard.secretsList',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          chromeModule.name,
          componentsModule.name,
          paginationModule.name,
          filtersModule.name,
        ])
    .config(stateConfig)
    .component('kdSecretCardList', secretCardListComponent)
    .component('kdSecretCard', secretCardComponent);
