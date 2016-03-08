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

import {Type} from './type';

/**
 * @final
 * @extends {Type}
 */
export class LabelKeyPrefixLengthType extends Type {
  constructor() { super(); }

  /**
   * Returns true if the label key prefix (before the "/" if there is one) is equal or shorter than
   * 253 characters, otherwise returns false.
   *
   * @override
   */

  isValid(value) {
    /** @type {!RegExp} */
    let PrefixPattern = /^(.*\/.*)$/;
    /** @type {boolean} */
    let isPrefixed = PrefixPattern.test(value);
    /** @type {number} */
    let slashPosition = isPrefixed ? value.indexOf("/") : -1;
    /** @type {number} */
    let maxKeyLength = 253;
    /** @type {string} */
    let labelKeyPrefix = isPrefixed ? value.substring(0, slashPosition) : '';

    return (labelKeyPrefix.length <= maxKeyLength);
  }
}
