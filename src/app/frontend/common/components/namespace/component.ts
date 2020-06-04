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

import {Component, ElementRef, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {MatDialog} from '@angular/material/dialog';
import {MatSelect} from '@angular/material/select';
import {ActivatedRoute, NavigationEnd, Router} from '@angular/router';
import {NamespaceList} from '@api/backendapi';
import {NamespaceDetail} from '@api/backendapi';
import {RoleBindingList} from '@api/backendapi';
import {PodList} from '@api/backendapi';
import {Subject} from 'rxjs';
import {startWith, switchMap, takeUntil, first} from 'rxjs/operators';

import {CONFIG} from '../../../index.config';
import {NAMESPACE_STATE_PARAM} from '../../params/params';
import {HistoryService} from '../../services/global/history';
import {NamespaceService} from '../../services/global/namespace';
import {NotificationSeverity, NotificationsService} from '../../services/global/notifications';
import {KdStateService} from '../../services/global/state';
import {EndpointManager, Resource} from '../../services/resource/endpoint';
import {ResourceService} from '../../services/resource/resource';

import {NamespaceChangeDialog} from './changedialog/dialog';
import {PermissionsService} from 'common/services/global/permissions';

@Component({
  selector: 'kd-namespace-selector',
  templateUrl: './template.html',
  styleUrls: ['style.scss'],
})
export class NamespaceSelectorComponent implements OnInit, OnDestroy {
  private namespaceUpdate_ = new Subject();
  private roleBindingUpdate_ = new Subject();
  private unsubscribe_ = new Subject();
  private readonly endpoint_ = EndpointManager.resource(Resource.namespace);
  private readonly endpointRole_ = EndpointManager.resource(Resource.roleBinding);

  namespaces: string[] = [];
  selectNamespaceInput = '';
  allNamespacesKey: string;
  selectedNamespace: string;
  resourceNamespaceParam: string;

  @ViewChild(MatSelect, {static: true}) private readonly select_: MatSelect;
  @ViewChild('namespaceInput', {static: true}) private readonly namespaceInputEl_: ElementRef;

  constructor(
    private readonly router_: Router,
    private readonly namespaceService_: NamespaceService,
    private readonly namespace_: ResourceService<NamespaceList>,
    private readonly roleBinding_: ResourceService<RoleBindingList>,
    private readonly podList_: ResourceService<PodList>,
    private readonly dialog_: MatDialog,
    private readonly kdState_: KdStateService,
    public permission: PermissionsService,
    private readonly notifications_: NotificationsService,
    private readonly _activatedRoute: ActivatedRoute,
    private readonly _historyService: HistoryService,
  ) {}

  ngOnInit(): void {
    this._activatedRoute.queryParams.pipe(takeUntil(this.unsubscribe_)).subscribe(params => {
      const namespace = params.namespace;
      if (!namespace) {
        this.setDefaultQueryParams_();
        return;
      }

      if (this.namespaceService_.current() === namespace) {
        return;
      }

      this.namespaceService_.setCurrent(namespace);
      this.namespaceService_.onNamespaceChangeEvent.emit(namespace);
      this.selectedNamespace = namespace;
    });

    this.resourceNamespaceParam = this._getCurrentResourceNamespaceParam();
    this.router_.events
      .filter(event => event instanceof NavigationEnd)
      .distinctUntilChanged()
      .subscribe(() => {
        this.resourceNamespaceParam = this._getCurrentResourceNamespaceParam();
        if (this.shouldShowNamespaceChangeDialog(this.namespaceService_.current())) {
          this.handleNamespaceChangeDialog_();
        }
      });

    this.allNamespacesKey = this.namespaceService_.getAllNamespacesKey();
    this.selectedNamespace = this.namespaceService_.current();
    this.select_.value = this.selectedNamespace;

    this.loadNamespaces_();
    //this.loadRoleBindings_();
  }

  ngOnDestroy(): void {
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
  }

  selectNamespace(): void {
    if (this.selectNamespaceInput.length > 0) {
      this.selectedNamespace = this.selectNamespaceInput;
      this.select_.close();
      this.changeNamespace_(this.selectedNamespace);
    }
  }

  onNamespaceToggle(opened: boolean): void {
    if (opened) {
      this.namespaceUpdate_.next();
      this.focusNamespaceInput_();
    } else {
      this.changeNamespace_(this.selectedNamespace);
    }
  }

  formatNamespaceName(namespace: string): string {
    if (this.namespaceService_.isMultiNamespace(namespace)) {
      return 'All namespaces';
    }

    return namespace;
  }

  /**
   * When state is loaded and namespaces are fetched perform basic validation.
   */
  private onNamespaceLoaded_(): void {
    let newNamespace = this.namespaceService_.getDefaultNamespace();
    const targetNamespace = this.selectedNamespace;

    if (
      targetNamespace &&
      (this.namespaces.indexOf(targetNamespace) >= 0 ||
        targetNamespace === this.allNamespacesKey ||
        this.namespaceService_.isNamespaceValid(targetNamespace))
    ) {
      newNamespace = targetNamespace;
    }

    if (newNamespace !== this.selectedNamespace) {
      this.changeNamespace_(newNamespace);
    }
  }

  private loadNamespaces_(): void {
    this.namespaceUpdate_
      .pipe(takeUntil(this.unsubscribe_))
      .pipe(startWith({}))
      .pipe(switchMap(() => this.namespace_.get(this.endpoint_.list())))
      .subscribe(
        namespaceList => {
          const namespacesTemp = namespaceList.namespaces.map(n => n.objectMeta.name);
          if (namespaceList.errors.length > 0) {
            for (const err of namespaceList.errors) {
              this.notifications_.pushErrors([err]);
            }
          }
          for (const namespace of namespacesTemp) {
            if (this.namespaces.indexOf(namespace) === -1) {
              this.checkNamespaces_(namespace);
            }
          }
        },
        () => {},
        () => {
          this.onNamespaceLoaded_();
        },
      );
  }

  private loadRoleBindings_(): void {
    this.roleBindingUpdate_
      .pipe(takeUntil(this.unsubscribe_))
      .pipe(startWith({}))
      .pipe(switchMap(() => this.roleBinding_.get(this.endpointRole_.list())))
      .subscribe(
        roleBindingList => {
          if (roleBindingList.errors.length > 0) {
            for (const err of roleBindingList.errors) {
              this.notifications_.pushErrors([err]);
              //console.log([err]);
            }
          }
        },
        () => {},
        () => {},
      );
  }

  private checkNamespaces_(namespaceName: string): void {
    this.podList_
      .get('api/v1/pod/' + namespaceName)
      .pipe(first())
      .subscribe(
        podList => {
          if (this.namespaces.indexOf(namespaceName) === -1) {
            if (podList.errors.length === 0) {
              //console.log(namespaceName + ' is allowed. Adding');
              this.namespaces.push(namespaceName);
            } else {
              const index = this.namespaces.indexOf(namespaceName, 0);
              if (index > -1) {
                //console.log(namespaceName + ' is forbidden. Deleting');
                this.namespaces.splice(index, 1);
              }
            }
          }
        },
        () => {},
        () => {},
      );
  }

  private handleNamespaceChangeDialog_(): void {
    this.dialog_
      .open(NamespaceChangeDialog, {
        data: {
          namespace: this.selectedNamespace,
          newNamespace: this._getCurrentResourceNamespaceParam(),
        },
      })
      .afterClosed()
      .subscribe(confirmed => {
        if (confirmed) {
          this.selectedNamespace = this._getCurrentResourceNamespaceParam();
          this.router_.navigate([], {
            relativeTo: this._activatedRoute,
            queryParams: {[NAMESPACE_STATE_PARAM]: this.selectedNamespace},
            queryParamsHandling: 'merge',
          });
        } else {
          this._historyService.goToPreviousState('overview');
        }
      });
  }

  private changeNamespace_(namespace: string): void {
    this.clearNamespaceInput_();

    if (this.resourceNamespaceParam) {
      // Go to overview of the new namespace as change was done from details view.
      this.router_.navigate(['overview'], {
        queryParams: {[NAMESPACE_STATE_PARAM]: namespace},
        queryParamsHandling: 'merge',
      });
    } else {
      // Change only the namespace as currently not on details view.
      this.router_.navigate([], {
        relativeTo: this._activatedRoute,
        queryParams: {[NAMESPACE_STATE_PARAM]: namespace},
        queryParamsHandling: 'merge',
      });
    }
  }

  private clearNamespaceInput_(): void {
    this.selectNamespaceInput = '';
  }

  private shouldShowNamespaceChangeDialog(targetNamespace: string): boolean {
    return (
      targetNamespace !== this.allNamespacesKey &&
      !!this.resourceNamespaceParam &&
      this.resourceNamespaceParam !== targetNamespace
    );
  }

  private _getCurrentResourceNamespaceParam(): string | undefined {
    return this._getCurrentRoute().snapshot.params.resourceNamespace;
  }

  private _getCurrentRoute(): ActivatedRoute {
    let route = this._activatedRoute.root;
    while (route && route.firstChild) {
      route = route.firstChild;
    }
    return route;
  }

  /**
   * Focuses namespace input field after clicking on namespace selector menu.
   */
  private focusNamespaceInput_(): void {
    // Wrap in a timeout to make sure that element is rendered before looking for it.
    setTimeout(() => {
      this.namespaceInputEl_.nativeElement.focus();
    }, 150);
  }

  setDefaultQueryParams_() {
    this.permission.redirectToNs(this.router_);
  }
}
