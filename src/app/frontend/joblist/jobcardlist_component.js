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

/**
 * @final
 */
class JobListController {
  /**
   * @param {!./../common/namespace/namespace_service.NamespaceService} kdNamespaceService
   * @ngInject
   */
  constructor(kdNamespaceService) {
    /** @private {!./../common/namespace/namespace_service.NamespaceService} */
    this.kdNamespaceService_ = kdNamespaceService;

    /** @export */
    this.i18n = i18n;
  }

  /**
   * @return {boolean}
   * @export
   */
  areMultipleNamespacesSelected() {
    return this.kdNamespaceService_.areMultipleNamespacesSelected();
  }
}

/**
 * @return {!angular.Component}
 */
export const jobCardListComponent = {
  transclude: true,
  bindings: {
    /** {!backendApi.JobList} */
    'jobList': '<',
    'jobListResource': '<',
    'showResourceKind': '<',
  },
  templateUrl: 'joblist/jobcardlist.html',
  controller: JobListController,
};

const i18n = {
  /** @export {string} @desc Label 'Name' which appears as a column label in the table of
   jobs (Job list view). */
  MSG_JOB_LIST_NAME_LABEL: goog.getMsg('Name'),
  /** @export {string} @desc Label 'Namespace' which appears as a column label in the
   table of replication controllers (Job list view). */
  MSG_JOB_LIST_NAMESPACE_LABEL: goog.getMsg('Namespace'),
  /** @export {string} @desc Label 'Labels' which appears as a column label in the table of
   replication controllers (Job list view). */
  MSG_JOB_LIST_LABELS_LABEL: goog.getMsg('Labels'),
  /** @export {string} @desc Label 'Pods' which appears as a column label in the table of
   replication controllers (Job list view). */
  MSG_JOB_LIST_PODS_LABEL: goog.getMsg('Pods'),
  /** @export {string} @desc Label 'Age' which appears as a column label in the
   table of replication controllers (Job list view). */
  MSG_JOB_LIST_AGE_LABEL: goog.getMsg('Age'),
  /** @export {string} @desc Label 'Images' which appears as a column label in the
   table of replication controllers (Job list view). */
  MSG_JOB_LIST_IMAGES_LABEL: goog.getMsg('Images'),
};
