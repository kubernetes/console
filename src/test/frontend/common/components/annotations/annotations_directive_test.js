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

import componentsModule from 'common/components/components_module';

describe('annotations directive', () => {
  /** @type {!angular.Scope} */
  let scope;
  /** @type {function(!angular.Scope):!angular.JQLite} */
  let compileFn;

  beforeEach(() => {
    angular.mock.module(componentsModule.name);

    angular.mock.inject(($rootScope, $compile) => {
      scope = $rootScope.$new();
      compileFn = $compile('<kd-annotations labels="annotations"></kd-annotations>');
    });
  });

  it('should render 3 annotations of unknown kind as labels', () => {
    // given
    scope.annotations = {
      app: 'app',
      version: 'version',
      testLabel: 'test',
    };

    // when
    let element = compileFn(scope);
    scope.$digest();

    // then
    let labels = element.find('kd-middle-ellipsis');
    expect(labels.length).toEqual(3);
    angular.forEach(scope.annotations, (value, key, index) => {
      expect(labels.eq(index).text()).toBe(`${key}=${value}`);
    });
  });

  it('should render 1 annotation of created-by kind as serialized reference', () => {
    // given
    scope.annotations = {
      app: 'app',
      'kubernetes.io/created-by': '{bogus: "json"}',
      testLabel: 'test',
    };

    // when
    let element = compileFn(scope);
    scope.$digest();

    // then
    let labels = element.find('kd-middle-ellipsis');
    expect(labels.length).toEqual(2);
    angular.forEach(scope.annotations, (value, key, index) => {
      expect(labels.eq(index).text()).toBe(`${key}=${value}`);
    });
    let annotations = element.find('kd-serialized-reference');
    expect(annotations.length).toEqual(1);
    expect(annotations.eq(0).text().trim()).toBe('{bogus: "json"}');

  });
});
