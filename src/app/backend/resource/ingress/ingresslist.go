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

package ingress

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// Ingress - a single ingress returned to the frontend.
type Ingress struct {
	common.ObjectMeta `json:"objectMeta"`
	common.TypeMeta   `json:"typeMeta"`

	// External endpoints of this ingress.
	Endpoints []common.Endpoint `json:"endpoints"`
}

// IngressList - response structure for a queried ingress list.
type IngressList struct {
	common.ListMeta `json:"listMeta"`

	// Unordered list of Ingresss.
	Items []Ingress `json:"items"`
}

// GetIngressList - return all ingresses in the given namespace.
func GetIngressList(client client.Interface, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*IngressList, error) {
	ingressList, err := client.Extensions().Ingress(namespace.ToRequestParam()).List(api.ListOptions{
		LabelSelector: labels.Everything(),
		FieldSelector: fields.Everything(),
	})
	if err != nil {
		return nil, err
	}
	return NewIngressList(ingressList.Items, dsQuery), err
}

// NewIngress - creates a new instance of Ingress struct based on K8s Ingress.
func NewIngress(ingress *extensions.Ingress) *Ingress {
	modelIngress := &Ingress{
		ObjectMeta: common.NewObjectMeta(ingress.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindIngress),
		Endpoints:  make([]common.Endpoint, 0),
	}

	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		for _, status := range ingress.Status.LoadBalancer.Ingress {
			endpoint := common.Endpoint{Host: status.IP}
			modelIngress.Endpoints = append(modelIngress.Endpoints, endpoint)
		}
	}

	return modelIngress
}

// NewIngressList - creates a new instance of IngressList struct based on K8s Ingresss array.
func NewIngressList(ingresses []extensions.Ingress, dsQuery *dataselect.DataSelectQuery) *IngressList {
	newIngressList := &IngressList{
		ListMeta: common.ListMeta{TotalItems: len(ingresses)},
		Items:    make([]Ingress, 0),
	}

	ingresses = fromCells(dataselect.GenericDataSelect(toCells(ingresses), dsQuery))

	for _, ingress := range ingresses {
		newIngressList.Items = append(newIngressList.Items, *NewIngress(&ingress))
	}

	return newIngressList
}
