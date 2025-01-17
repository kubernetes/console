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

import {OverlayContainer} from '@angular/cdk/overlay';
import {Component, ElementRef, OnInit} from '@angular/core';

import {LocalSettingsService} from '@common/services/global/localsettings';
import {ThemeService} from '@common/services/global/theme';
import {TimezoneService} from '@common/services/global/timezone';
import {TitleService} from '@common/services/global/title';

@Component({selector: 'kd-root', template: '<router-outlet></router-outlet>'})
export class RootComponent implements OnInit {
  private _theme = this._themeService.theme;
  private _timezone = this._timezoneService.timezone;

  constructor(
    private readonly _themeService: ThemeService,
    private readonly _localSettingService: LocalSettingsService,
    private readonly _overlayContainer: OverlayContainer,
    private readonly _kdRootRef: ElementRef,
    private readonly _titleService: TitleService,
    private readonly _timezoneService: TimezoneService
  ) {}

  ngOnInit(): void {
    this._titleService.update();
    this._themeService.subscribe(this.onThemeChange_.bind(this));

    const localSettings = this._localSettingService.get();
    if (localSettings) {
      if (localSettings.theme) {
        this._theme = localSettings.theme;
        this._themeService.theme = localSettings.theme;
      }
      if (localSettings.timezone) {
        this._timezone = localSettings.timezone;
        this._timezoneService.timezone = localSettings.timezone;
      }
    }

    this.applyOverlayContainerTheme_('', this._theme);
  }

  private onThemeChange_(theme: string): void {
    this.applyOverlayContainerTheme_(this._theme, theme);
    this._theme = theme;
  }

  private applyOverlayContainerTheme_(oldTheme: string, newTheme: string): void {
    if (!!oldTheme && oldTheme !== newTheme) {
      this._overlayContainer.getContainerElement().classList.remove(oldTheme);
      this._kdRootRef.nativeElement.classList.remove(oldTheme);
    }

    this._overlayContainer.getContainerElement().classList.add(newTheme);
    this._kdRootRef.nativeElement.classList.add(newTheme);
  }
}
