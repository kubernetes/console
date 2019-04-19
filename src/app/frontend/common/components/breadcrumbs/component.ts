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
import {ActivatedRoute, NavigationEnd, Params, Route, Router} from '@angular/router';
import {Breadcrumb} from '@api/frontendapi';
import {POD_DETAIL_ROUTE} from '../../../resource/workloads/pod/routing';

export const LOGS_PARENT_PLACEHOLDER = '___LOGS_PARENT_PLACEHOLDER___';

@Component({
  selector: 'kd-breadcrumbs',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
})
export class BreadcrumbsComponent implements OnInit {
  breadcrumbs: Breadcrumb[];

  constructor(private readonly _router: Router, private readonly _activatedRoute: ActivatedRoute) {}

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
    const currentRoute = this._getCurrentRoute();
    let url = '';

    this.breadcrumbs = [{
      label: this._getBreadcrumbLabel(currentRoute.routeConfig, currentRoute.snapshot.params),
      stateLink: url,
    }];

    let route: Route;
    if (currentRoute && currentRoute.routeConfig && currentRoute.routeConfig.data &&
        currentRoute.routeConfig.data.parent) {
      if (currentRoute.routeConfig.data.parent === LOGS_PARENT_PLACEHOLDER) {
        route = this._getLogsParent(currentRoute.snapshot.params);
      } else {
        route = currentRoute.routeConfig.data.parent;
      }

      while (route) {
        url = `/${route.path}/${url}`;  // TODO

        this.breadcrumbs.push({
          label: this._getBreadcrumbLabel(route, currentRoute.snapshot.params),
          stateLink: url,
        });

        // Explore the route tree to the root route (parent references have to be defined by us on
        // each route).
        if (route && route.data && route.data.parent) {
          route = route.data.parent;
        } else {
          break;
        }
      }
    }

    this.breadcrumbs.reverse();
  }

  private _getLogsParent(params: Params): Route|undefined {
    const resourceType = params['resourceType'];
    switch (resourceType) {
      case 'pod':
        return POD_DETAIL_ROUTE;
      default:
        return undefined;
    }
  }

  private _getCurrentRoute(): ActivatedRoute {
    let route = this._activatedRoute.root;
    while (route && route.firstChild) {
      route = route.firstChild;
    }
    return route;
  }

  // TODO: Add data to all structures.
  // TODO: When state search is active use specific logic to display custom breadcrumb:
  //  if (state.url[0].path === searchState.name) {
  //    const query = stateParams[SEARCH_QUERY_STATE_PARAM];
  //    return `Search for "${query}"`;
  //  }
  private _getBreadcrumbLabel(route: Route, params: Params) {
    if (route && route.data && route.data.breadcrumb) {
      let breadcrumb = route.data.breadcrumb as string;
      if (breadcrumb.startsWith('{{') && breadcrumb.endsWith('}}')) {
        breadcrumb = breadcrumb.slice(2, breadcrumb.length - 2).trim();
        breadcrumb = params[breadcrumb];
      }
      return breadcrumb;
    } else if (route && route.component) {
      return route.component.name;
    } else {
      return 'Unknown';
    }
  }
}
