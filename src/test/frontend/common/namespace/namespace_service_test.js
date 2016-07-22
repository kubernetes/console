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

import namespaceModule from 'common/namespace/namespace_module';

describe('Namespace service', () => {
  /** @type {!common/namespace/namespace_service.NamespaceService} */
  let namespaceService;

  beforeEach(() => angular.mock.module(namespaceModule.name));

  beforeEach(angular.mock.inject((kdNamespaceService) => {
    namespaceService = kdNamespaceService;

  }));

  it(`should initialise multipleNamespacesSelected as true`,
     () => { expect(namespaceService.getMultipleNamespacesSelected()).toBe(true); });

  it(`should set multipleNamespacesSelected to true`, () => {
    namespaceService.setMultipleNamespacesSelected(true);
    expect(namespaceService.getMultipleNamespacesSelected()).toBe(true);
  });

  it(`should set multipleNamespacesSelected to false`, () => {
    namespaceService.setMultipleNamespacesSelected(false);
    expect(namespaceService.getMultipleNamespacesSelected()).toBe(false);
  });
});
