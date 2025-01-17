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

import {Timezone} from '@api/root.api';
import {getAllTimezones, getTimezone} from 'countries-and-timezones';

export class TimezoneService {
  private _timezones: Timezone[] = [];
  static DefaultTimezone: string = Intl.DateTimeFormat().resolvedOptions().timeZone;
  private _timezone = TimezoneService.DefaultTimezone;
  constructor() {}

  get timezone(): string {
    return this._timezone;
  }

  set timezone(timezone: string) {
    this._timezone = timezone;
  }

  get timezones(): Timezone[] {
    return this._timezones;
  }

  get utcOffset(): string {
    return getTimezone(this._timezone).utcOffsetStr;
  }

  init(): void {
    this._timezones = Object.entries(getAllTimezones())
      .sort((a, b) => a[1].utcOffset - b[1].utcOffset)
      .map(([key, value]) => ({
        label: `${key} (UTC ${value.utcOffsetStr})`,
        value: key,
      }));
  }
}
