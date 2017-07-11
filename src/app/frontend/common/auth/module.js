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

import {AuthService} from './service';

import {stateName as loginState} from 'login/state';

/**
 * Angular module containing application configuration.
 */
export default angular
    .module(
        'kubernetesDashboard.auth',
        [
          'ngCookies',
          'ngResource',
        ])
    .service('kdAuthService', AuthService)
    .run(initAuthService);

/**
 * Initializes the service to track state changes and make sure that user is logged in and
 * token has not expired.
 */
/**
 * @ngInject
 */
function initAuthService(kdAuthService, $rootScope, $state) {
  $rootScope.$on('$stateChangeStart', (event, toState) => {
    if (toState.name !== loginState && !kdAuthService.isLoggedIn() &&
        kdAuthService.isLoginPageEnabled() && !kdAuthService.isAuthHeaderPresent()) {
      event.preventDefault();
      return $state.transitionTo(loginState);
    }
  });
}
