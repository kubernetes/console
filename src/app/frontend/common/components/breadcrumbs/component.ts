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

import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, ActivatedRouteSnapshot, NavigationEnd, Router} from '@angular/router';
import {Breadcrumb} from '@api/frontendapi';
import {BreadcrumbsService} from '../../services/global/breadcrumbs';

@Component({
  selector: 'kd-breadcrumbs',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class BreadcrumbsComponent implements OnInit {
  breadcrumbs: Breadcrumb[] = [];

  constructor(
      private readonly _router: Router, private readonly _activatedRoute: ActivatedRoute,
      private readonly _breadcrumbs: BreadcrumbsService) {}

  ngOnInit(): void {
    this._registerNavigationHook();
  }

  private _registerNavigationHook(): void {
    this._router.events.filter(event => event instanceof NavigationEnd)
        .distinctUntilChanged()
        .subscribe(() => {
          this._initBreadcrumbs();
        });
  }

  private _initBreadcrumbs(): void {
    this.breadcrumbs = [];

    const currentRoute = this._getCurrentRoute();

    // TODO
    this.breadcrumbs.push({
      label: currentRoute.routeConfig.component.name,
      stateLink: '',
    });

    // const label = route.routeConfig && route.routeConfig.data ? route.routeConfig.data[
    //   'breadcrumb' ] : 'Home'; const nextUrl = `${url}${route.routeConfig ?
    //   route.routeConfig.path :
    //   ''}/`; breadcrumbs.push({label: label, stateLink: nextUrl}); return route.firstChild ?
    //   this._buildBreadCrumb(route.firstChild, nextUrl, breadcrumbs) : breadcrumbs;



    // _buildBreadCrumb(route: ActivatedRoute, url: string = '', breadcrumbs: Breadcrumb[] = []):
    // Breadcrumb[] {
    //   const label = route.routeConfig && route.routeConfig.data ? route.routeConfig.data[
    //   'breadcrumb' ] : 'Home'; const nextUrl = `${url}${route.routeConfig ?
    //   route.routeConfig.path :
    //   ''}/`; breadcrumbs.push({label: label, stateLink: nextUrl}); return route.firstChild ?
    //   this._buildBreadCrumb(route.firstChild, nextUrl, breadcrumbs) : breadcrumbs;
    // }



    // let state: ActivatedRouteSnapshot = this.router_.routerState.root.snapshot;
    // const breadcrumbs: Breadcrumb[] = [];
    //
    // while (state && state.url && this.canAddBreadcrumb_(breadcrumbs)) {
    //   const breadcrumb = this.getBreadcrumb_(state);
    //
    //   if (breadcrumb.label) {
    //     breadcrumbs.push(breadcrumb);
    //   }
    //
    //   state = this.breadcrumbs_.getParentState(state).snapshot;
    // }
    //
    // this.breadcrumbs = breadcrumbs.reverse();
  }

  private _getCurrentRoute(): ActivatedRoute {
    let route = this._activatedRoute.root;
    while (route && route.firstChild) {
      route = route.firstChild;
    }
    return route;
  }

  private getBreadcrumb_(state: ActivatedRouteSnapshot): Breadcrumb {
    const breadcrumb = new Breadcrumb();

    if (state) {
    }
    // breadcrumb.label = this.breadcrumbs_.getDisplayName(state);
    // breadcrumb.stateLink = this.router_.href(state.name, this.router_.params);

    return breadcrumb;
  }
}
