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

import path from 'path';

import DeployFromFilePageObject from '../deploy/deployfromfile_po';

import remote from '../../../../node_modules/selenium-webdriver/remote';

// Test assumes, that there are no replication controllers in the cluster at the beginning.
// TODO(#494): Reenable this test when fixed.
describe('Deploy from invalid file user story test', () => {

  /** @type {!DeployFromFilePageObject} */
  let deployFromFilePage;

  beforeAll(() => {
    browser.driver.setFileDetector(new remote.FileDetector());
    deployFromFilePage = new DeployFromFilePageObject();
    browser.get('#/deploy');
    // switches to deploy from file
    deployFromFilePage.deployFromFileRadioButton.click();
  });

  it('should pop up error dialog after uploading the invalid file', () => {
    // given
    let fileToUpload = '../deploy/invalid-rc.yaml';
    let absolutePath = path.resolve(__dirname, fileToUpload);

    // when
    browser.driver.executeScript(deployFromFilePage.makeInputVisible);
    deployFromFilePage.filePicker.sendKeys(absolutePath);
    deployFromFilePage.deployButton.click();

    // then
    expect(deployFromFilePage.mdDialog.isPresent()).toBeTruthy();
    expect(browser.getCurrentUrl()).toContain('deploy');
  });
});
