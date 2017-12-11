// Copyright 2017 The Kubernetes Authors.
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

export default class DeployFromFilePageObject {
  constructor() {
    this.deployFromFileTabQuery = by.xpath('//md-tab-item[contains(text(),"from file")]');
    this.deployFromFileTab = element(this.deployFromFileTabQuery);

    this.deployButtonQuery = by.css('.kd-deploy-from-file-button');
    this.deployButton = element(this.deployButtonQuery);

    this.filePickerQuery = by.css('.kd-upload-button');
    this.filePicker_ = element(this.filePickerQuery);

    this.mdDialogQuery = by.tagName('md-dialog');
    this.mdDialog = element(this.mdDialogQuery);
  }

  /**
   * Sets filepath on the filePicker input field
   * @param {string} filePath
   */
  setFile(filePath) {
    this.filePicker_.sendKeys(filePath);
  }
}
