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

import {Component, Input} from '@angular/core';
import {MatTableDataSource} from '@angular/material/table';
import {PersistentVolumeSource} from '@api/volume.api';
import {StringMap} from '@api/root.shared';

@Component({
  selector: 'kd-persistent-volume-source',
  templateUrl: './template.html',
})
export class PersistentVolumeSourceComponent {
  @Input() source: PersistentVolumeSource;
  @Input() initialized: boolean;

  getVolumeAttributesColumns(): string[] {
    return ['key', 'value'];
  }

  getVolumeAttributesDataSource(): MatTableDataSource<StringMap> {
    const data: StringMap[] = [];

    if (this.initialized) {
      for (const rName of Array.from<string>(Object.keys(this.source.csi.volumeAttributes))) {
        data.push({
          key: rName,
          value: this.source.csi.volumeAttributes[rName],
        });
      }
    }

    const tableData = new MatTableDataSource<StringMap>();
    tableData.data = data;

    return tableData;
  }

  trackByCapacityItemName(_: number, item: StringMap): string {
    return item.key;
  }
}
