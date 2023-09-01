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

package ingressroute

import (
	"context"
	"fmt"
	"time"

	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/dashboard/api/pkg/api"
	"k8s.io/dashboard/api/pkg/errors"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
)

// Ingress - a single ingress returned to the frontend.
type Ingress struct {
	api.ObjectMeta `json:"objectMeta"`
	api.TypeMeta   `json:"typeMeta"`

	// External endpoints of this ingress.
	Endpoints []common.Endpoint `json:"endpoints"`
	Hosts     []string          `json:"hosts"`
}

type IngressRoute struct {
	api.ObjectMeta `json:"objectMeta"`
	api.TypeMeta   `json:"typeMeta"`

	// Host of this ingress route.
	Entrypoints []string `json:"entrypoints"`
	Hosts      []string `json:"hosts"`
}

// IngressList - response structure for a queried ingress list.
type IngressList struct {
	api.ListMeta `json:"listMeta"`

	// Unordered list of Ingresss.
	Items []Ingress `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}


type IngressRouteList struct {
	api.ListMeta `json:"listMeta"`

	// Unordered list of Ingressroutes.
	Items []IngressRoute `json:"items"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetIngressList returns all ingresses in the given namespace.
func GetIngressRouteList(client client.Interface, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery, config *rest.Config) (*IngressRouteList, error) {
	_, err := client.NetworkingV1().Ingresses(namespace.ToRequestParam()).List(context.TODO(), api.ListEverything)
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	traefikclient, err := traefik.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	ingressList, err := traefikclient.TraefikV1alpha1().IngressRoutes("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
					// handle err
	}
	ingressCtrls := ingressList.Items
	if len(ingressCtrls) > 0 {
					for _, ingress := range ingressCtrls {
									fmt.Printf("ingress %s exists in namespace %s\n", ingress.Name, ingress.Namespace)
					}
	} else {
					fmt.Println("no ingress found")
	}
	time.Sleep(10 * time.Second)

	// //nonCriticalErrors, criticalError := errors.HandleError(err)
	// if criticalError != nil {
	// 	return nil, criticalError
	// }

	return nil, nil
}

// GetIngressList returns all ingresses in the given namespace.
func GetIngressList(client client.Interface, namespace *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*IngressList, error) {
	ingressList, err := client.NetworkingV1().Ingresses(namespace.ToRequestParam()).List(context.TODO(), api.ListEverything)
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToIngressList(ingressList.Items, nonCriticalErrors, dsQuery), nil
}

func getEndpoints(ingress *v1.Ingress) []common.Endpoint {
	endpoints := make([]common.Endpoint, 0)
	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		for _, status := range ingress.Status.LoadBalancer.Ingress {
			endpoint := common.Endpoint{}
			if status.Hostname != "" {
				endpoint.Host = status.Hostname
			} else if status.IP != "" {
				endpoint.Host = status.IP
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func getHosts(ingress *v1.Ingress) []string {
	hosts := make([]string, 0)
	set := make(map[string]struct{})

	for _, rule := range ingress.Spec.Rules {
		if _, exists := set[rule.Host]; !exists && len(rule.Host) > 0 {
			hosts = append(hosts, rule.Host)
		}

		set[rule.Host] = struct{}{}
	}

	return hosts
}

func toIngress(ingress *v1.Ingress) Ingress {
	return Ingress{
		ObjectMeta: api.NewObjectMeta(ingress.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindIngress),
		Endpoints:  getEndpoints(ingress),
		Hosts:      getHosts(ingress),
	}
}

func ToIngressList(ingresses []v1.Ingress, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *IngressList {
	newIngressList := &IngressList{
		ListMeta: api.ListMeta{TotalItems: len(ingresses)},
		Items:    make([]Ingress, 0),
		Errors:   nonCriticalErrors,
	}

	ingresCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(ingresses), dsQuery)
	ingresses = fromCells(ingresCells)
	newIngressList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, ingress := range ingresses {
		newIngressList.Items = append(newIngressList.Items, toIngress(&ingress))
	}

	return newIngressList
}
