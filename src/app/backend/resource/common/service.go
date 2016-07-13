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

package common

import (
	"k8s.io/kubernetes/pkg/api"
)

// TODO add doc
func FilterNamespacedServicesBySelector(services []api.Service, namespace string,
	resourceSelector map[string]string) []api.Service {

	var matchingServices []api.Service
	for _, service := range services {
		if service.ObjectMeta.Namespace == namespace &&
			IsSelectorMatching(resourceSelector, service.Labels) {
			matchingServices = append(matchingServices, service)
		}
	}

	return matchingServices
}
