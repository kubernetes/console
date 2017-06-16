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

import namespaceModule from 'namespace/module';

describe('Namespace card', () => {
  /**
   * @type {!NamespaceCardController} */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(namespaceModule.name);

    angular.mock.inject(($componentController, $rootScope) => {
      ctrl = $componentController('kdNamespaceCard', {$scope: $rootScope});
    });
  });

  it('should construct details href', () => {
    // given
    ctrl.namespace = {
      objectMeta: {
        name: 'foo-name',
      },
    };

    // then
    expect(ctrl.getNamespaceDetailHref()).toEqual('#!/namespace/foo-name');
  });

  it('should format the "created at" tooltip correctly', () => {
    expect(ctrl.getCreatedAtTooltip('2016-06-06T09:13:12Z'))
        .toMatch('Created at 2016-06-06T09:13.*');
  });
});
