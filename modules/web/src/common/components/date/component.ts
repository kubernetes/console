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

import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  DestroyRef,
  inject,
  Input,
  OnChanges,
} from '@angular/core';
import {Subject, timer} from 'rxjs';
import {switchMap} from 'rxjs/operators';
import {takeUntilDestroyed} from '@angular/core/rxjs-interop';
import {TimezoneService} from '@common/services/global/timezone';

/**
 * Display a date
 *
 * Examples:
 *
 * Display the date:
 * <kd-date [date]="object.timestamp"></kd-date>
 *
 * Display the age of the date, and the date in a tooltip:
 * <kd-date [date]="object.timestamp" relative></kd-date>
 *
 * Display the date in the shprt format:
 * <kd-date [date]="object.timestamp" format="short"></kd-date>
 *
 * Display the age of the date, and the date in the short format in a tooltip:
 * <kd-date [date]="object.timestamp" relative format="short"></kd-date>
 *
 */
@Component({
  selector: 'kd-date',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DateComponent implements OnChanges {
  @Input() date: string;
  @Input() format = 'medium';

  @Input()
  set relative(v: boolean) {
    this.relative_ = v !== undefined && v !== false;
  }

  private destroyRef = inject(DestroyRef);

  get relative(): boolean {
    return this.relative_;
  }

  get timezone(): string {
    return this._timezone.utcOffset;
  }

  iteration = 0;

  private relative_: boolean;
  private refreshInterval_: number;
  private intervalChanged_ = new Subject<void>();
  private timeBaseIntervals_ = [
    60, // Seconds in a minute
    60, // Minutes in a hour
    24, // Hours in a day
  ];

  constructor(
    private readonly cdr_: ChangeDetectorRef,
    private readonly _timezone: TimezoneService
  ) {}

  ngOnChanges() {
    if (this.relative_) {
      this.intervalChanged_
        .pipe(switchMap(_ => timer(0, this.refreshInterval_)))
        .pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe(_ => {
          this.cdr_.markForCheck();
          this.iteration++;

          // Check if refresh interval should be updated
          const interval = this.calculateRefreshInterval_(this.getTimePassed_());
          if (interval !== this.refreshInterval_ / 1000) {
            this.setRefreshInterval_(interval);
          }
        });

      // Kick off the check interval
      this.setRefreshInterval_(1);
    }
  }

  // Calculates timer refresh interval that should be used based on time that has passed.
  // 1 - 59s -> 1 second
  // 1m - 59m59s -> 1 minute
  // 1h - 23h59m59s -> 1 hour
  // > 1 day -> 1 day
  private calculateRefreshInterval_(passed: number): number {
    let power = 0;
    let interval = 1;

    while (power < this.timeBaseIntervals_.length && passed / this.timeBaseIntervals_[power] >= 1) {
      passed /= this.timeBaseIntervals_[power];
      power++;
    }

    while (power > 0) {
      interval *= this.timeBaseIntervals_[--power];
    }

    return interval;
  }

  // Returns how much time has passed (in seconds) between the provided date and current time.
  private getTimePassed_(): number {
    return Math.floor((new Date().getTime() - new Date(this.date).getTime()) / 1000);
  }

  // Takes the interval in seconds and updates currently running timer to use the new interval
  private setRefreshInterval_(interval: number): void {
    this.refreshInterval_ = interval * 1000;
    this.intervalChanged_.next();
  }
}
