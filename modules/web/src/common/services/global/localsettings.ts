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

import {Injectable} from '@angular/core';

import {LocalSettings} from '@api/root.api';
import {ThemeService} from './theme';
import {TimezoneService} from './timezone';

@Injectable()
export class LocalSettingsService {
  private readonly _settingsKey = 'localSettings';
  private settings_: LocalSettings = {
    theme: ThemeService.SystemTheme,
    timezone: TimezoneService.DefaultTimezone,
  };

  constructor(
    private readonly theme_: ThemeService,
    private readonly timezone_: TimezoneService
  ) {}

  init(): void {
    const cookieValue = localStorage.getItem(this._settingsKey);
    if (cookieValue && cookieValue.length > 0) {
      this.settings_ = JSON.parse(cookieValue);
    }
  }

  get(): LocalSettings {
    return this.settings_;
  }

  handleThemeChange(theme: string): void {
    this.settings_.theme = theme;
    this.theme_.theme = theme;
    this.updateCookie_();
  }

  handleTimezoneChange(timezone: string): void {
    this.settings_.timezone = timezone;
    this.timezone_.timezone = timezone;
    this.updateCookie_();
  }

  updateCookie_(): void {
    localStorage.setItem(this._settingsKey, JSON.stringify(this.settings_));
  }
}
