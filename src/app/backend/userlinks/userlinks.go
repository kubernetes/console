// Copyright 2017 The Kubernetes Dashboard Authors.
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

package userlinks

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "k8s.io/client-go/kubernetes"
)

const (
	annotationObj     = "alpha.dashboard.kubernetes.io/links"
	apiserverProxyURL = "http://{{apiserver-proxy-url}}"
)

// UserLink is an optional annotation attached to the pod or service resource objects.
type UserLink struct {
	// Description is the key specified in the annotationObj set
	Description string `json:"description"`
	// Link is the value specified in the annotationObj set
	Link string `json:"link"`
	// Link status
	Valid bool `json:"valid"`
}

// GetUserLinks delegates getting of user links based on passed in resource kind (ResourceKindService, ResourceKindPod)
func GetUserLinks(client k8sClient.Interface, namespace, name, resource, host string) (userLinks []UserLink, err error) {
	log.Printf("Getting %s resource in %s namespace", name, namespace)

	switch {
	case resource == api.ResourceKindService:
		return getServiceLinks(client, namespace, name, host)
	case resource == api.ResourceKindPod:
		return getPodLinks(client, namespace, name, host)
	default:
		log.Printf("Unknown resource types %T!\n", resource)
	}
	return
}

// getServiceLinks get userlinks for services
func getServiceLinks(client k8sClient.Interface, namespace, name, host string) ([]UserLink, error) {
	userLinks := []UserLink{}
	service, err := client.CoreV1().Services(namespace).Get(name, metaV1.GetOptions{})
	if err != nil || len(service.Annotations[annotationObj]) == 0 {
		return userLinks, err
	}
	m := map[string]string{}
	err = json.Unmarshal([]byte(service.Annotations[annotationObj]), &m)
	if err != nil {
		return userLinks, err
	}

	for key, uri := range m {
		userLink := new(UserLink)
		userLink.Description = key
		if strings.Contains(uri, apiserverProxyURL) {
			userLink.Link = host + "/api/v1/namespaces/" + service.ObjectMeta.Namespace +
				"/services/" + service.ObjectMeta.Name + "/proxy/" + strings.TrimLeft(uri, apiserverProxyURL)
			userLink.Valid = true
		} else if _, err := url.ParseRequestURI(uri); err != nil {
			userLink.Link = "Invalid User Link: " + uri
			userLink.Valid = false
		} else {
			if len(service.Status.LoadBalancer.Ingress) > 0 {
				ingress := service.Status.LoadBalancer.Ingress[0]
				ip := ingress.IP
				if ip == "" {
					ip = ingress.Hostname
				}
				for _, port := range service.Spec.Ports {
					userLink.Link += "http://" + ip + ":" + strconv.Itoa(int(port.Port)) + uri
					userLink.Valid = true
				}
			} else {
				userLink.Link = uri
				userLink.Valid = true
			}
		}
		userLinks = append(userLinks, *userLink)
	}
	return userLinks, err
}

// getPodLinks get userlinks for links
func getPodLinks(client k8sClient.Interface, namespace, name, host string) ([]UserLink, error) {
	userLinks := []UserLink{}
	pod, err := client.CoreV1().Pods(namespace).Get(name, metaV1.GetOptions{})
	if err != nil || len(pod.Annotations[annotationObj]) == 0 {
		return userLinks, err
	}

	m := map[string]string{}
	err = json.Unmarshal([]byte(pod.Annotations[annotationObj]), &m)
	if err != nil {
		return userLinks, err
	}

	for key, uri := range m {
		userLink := new(UserLink)
		userLink.Description = key
		if strings.Contains(uri, apiserverProxyURL) {
			userLink.Link = host + "/api/v1/namespaces/" + pod.ObjectMeta.Namespace +
				"/pods/" + pod.ObjectMeta.Name + "/proxy/" + strings.TrimLeft(uri, apiserverProxyURL)
			userLink.Valid = true
		} else if _, err := url.ParseRequestURI(uri); err != nil {
			userLink.Link = "Invalid User Link: " + uri
			userLink.Valid = false
		} else {
			userLink.Link = uri
			userLink.Valid = true
		}
		userLinks = append(userLinks, *userLink)
	}
	return userLinks, err
}
