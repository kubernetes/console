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

import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {AppConfig} from '@api/root.api';
import {VersionInfo} from '@api/root.ui';
import {Observable} from 'rxjs';
import {version} from '@environments/version';

@Injectable()
export class ConfigService {
  private readonly configPath_ = 'config';
  private config_: AppConfig;
  private initTime_: number;
  private customConfig_: any;
  constructor(private readonly http: HttpClient) {}

  init(): void {
    this.getAppConfig().subscribe(config => {
      // Set init time when response from the backend will arrive.
      this.config_ = config;
      this.initTime_ = new Date().getTime();
    });
    this.readJson().subscribe(config => {
      this.customConfig_ = config;
      if (!this.customConfig_) {
        window.location.reload();
      }
    });
  }

  getAppConfig(): Observable<AppConfig> {
    return this.http.get<AppConfig>(this.configPath_);
  }

  getServerTime(): Date {
    if (this.config_.serverTime) {
      const elapsed = new Date().getTime() - this.initTime_;
      return new Date(this.config_.serverTime + elapsed);
    }
    return new Date();
  }

  getVersionInfo(): VersionInfo {
    return version;
  }

  readJson(): Observable<{}> {
    return this.http.get('assets/config/config.json');
  }
  getCustomConfig(): any {
    return this.customConfig_;
  }
  getColor(): string {
    if (this.customConfig_) {
      return this.customConfig_['color'];
    } else {
      return 'blue';
    }
  }
  getFileList(): string[] {
    if (this.customConfig_ !== undefined) {
      console.log(this.customConfig_['opaFiles']);
      return this.customConfig_['opaFiles'];
    } else {
      return [];
    }
  }
  getTitle(): string {
    if (this.customConfig_ !== undefined) {
      return this.customConfig_['title'];
    } else {
      return 'error';
    }
  }
  getMenus(): JSON {
    if (this.customConfig_ !== undefined) {
      return this.customConfig_['menus'];
    } else {
      return null;
    }
  }
}
