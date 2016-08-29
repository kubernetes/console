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

import {formatCpuUsage, formatMemoryUsage, formatTime} from './graph_tick_formatters';

const i18n = {
  /** @export {string} @desc Name of the CPU usage metric as displayed in the legend. */
  MSG_GRAPH_CPU_USAGE_LEGEND_LABEL: goog.getMsg('CPU Usage'),
  /** @export {string} @desc Name of the memory usage metric as displayed in the legend. */
  MSG_GRAPH_MEMORY_USAGE_LEGEND_LABEL: goog.getMsg('Memory Usage'),
  /** @export {string} @desc Name of the CPU limit metric as displayed in the legend. */
  MSG_GRAPH_CPU_LIMIT_LEGEND_LABEL: goog.getMsg('CPU Limit'),
  /** @export {string} @desc Name of Y axis showing CPU usage. */
  MSG_GRAPH_CPU_AXIS_LABEL: goog.getMsg('CPU (cores)'),
  /** @export {string} @desc Name of Y axis showing memory usage. */
  MSG_GRAPH_MEMORY_AXIS_LABEL: goog.getMsg('Memory (bytes)'),
  /** @export {string} @desc Name of time axis. */
  MSG_GRAPH_TIME_AXIS_LABEL: goog.getMsg('Time'),
};

export const CPUAxisType = 'CPUAxisType';
export const MemoryAxisType = 'MemoryAxisType';
export const TimeAxisType = 'TimeAxisType';

/**
 * Settings used by GraphController to display different metrics.
 *
 * @type {!Object<string, !Object<string, ?>>}
 */
export const metricDisplaySettings = {
  'cpu/usage_rate': {
    yAxisType: CPUAxisType,
    area: true,
    key: i18n.MSG_GRAPH_CPU_USAGE_LEGEND_LABEL,
    color: '#00c752',  // $chart-1
    fillOpacity: 0.2,
    strokeWidth: 3,
    type: 'line',
    yAxis: 1,
  },
  'cpu/limit': {
    yAxisType: CPUAxisType,
    area: true,
    key: i18n.MSG_GRAPH_CPU_LIMIT_LEGEND_LABEL,
    color: '#f51200',
    fillOpacity: 0.2,
    strokeWidth: 3,
    type: 'line',
    yAxis: 1,
  },
  'memory/usage': {
    yAxisType: MemoryAxisType,
    area: true,
    key: i18n.MSG_GRAPH_MEMORY_USAGE_LEGEND_LABEL,
    color: '#326de6',  // $chart-2
    fillOpacity: 0.2,
    strokeWidth: 3,
    type: 'line',
    yAxis: 2,
  },
};

/**
 * Settings used by GraphController to display different axes.
 *
 * @type {!Object<string, !Object<string, ?>>}
 */
export const axisSettings = {
  CPUAxisType: {
    formatter: formatCpuUsage,
    label: i18n.MSG_GRAPH_CPU_AXIS_LABEL,
  },
  MemoryAxisType: {
    formatter: formatMemoryUsage,
    label: i18n.MSG_GRAPH_MEMORY_AXIS_LABEL,
  },
  TimeAxisType: {
    formatter: formatTime,
    label: i18n.MSG_GRAPH_TIME_AXIS_LABEL,
  },
};
