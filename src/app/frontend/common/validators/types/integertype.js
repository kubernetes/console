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
 * @extends Type
 */
export class IntegerType extends Type {
  /**
   * @constructs IntegerType
   */
  constructor() { super(); }

  /**
   * Returns true if given value is a correct integer value, false otherwise.
   * When value is undefined or empty then it is considered as correct value in order
   * to not conflict with other validations like 'required'.
   *
   * @method
   * @param {number} value
   * @return {boolean}
   */
  isValid(value) { return (Number(value) === value && value % 1 === 0) || !value; }
}
