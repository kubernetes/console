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

import {GlobalSettings, K8sError} from '@api/backendapi';

export interface BreadcrumbConfig {
  label?: string;
  parent?: string;
}

export class Breadcrumb {
  label: string;
  stateLink: string;
}

export type ThemeSwitchCallback = (isLightThemeEnabled: boolean) => void;

export type onSettingsLoadCallback = (settings?: GlobalSettings) => void;
export type onSettingsFailCallback = (err?: KdError|K8sError) => void;

export interface KnownErrors { unauthorized: KdError; }

export interface KdError {
  status: string;
  code: number;
  message: string;
}

export interface OnListChangeEvent {
  id: string;
  groupId: string;
  items: number;
  filtered: boolean;
}

export interface ActionColumn {
  name: string;
  icon: string;
}
