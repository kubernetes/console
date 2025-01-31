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

package configmap

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// ConfigMapDetail API resource provides mechanisms to inject containers with configuration data while keeping
// containers agnostic of Kubernetes
type ConfigMapDetail struct {
	// Extends list item structure.
	ConfigMap `json:",inline"`

	// Data contains the configuration data.
	// Each key must be a valid DNS_SUBDOMAIN with an optional leading dot.
	Data map[string]string `json:"data,omitempty"`

	// BinaryData contains the configuration binary data.
	// Each key must be a valid DNS_SUBDOMAIN with an optional leading dot.
	BinaryData map[string]int `json:"binaryData,omitempty"`
}

// GetConfigMapDetail returns detailed information about a config map
func GetConfigMapDetail(client kubernetes.Interface, namespace, name string) (*ConfigMapDetail, error) {
	klog.V(4).Infof("Getting details of %s config map in %s namespace", name, namespace)

	rawConfigMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metaV1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return getConfigMapDetail(rawConfigMap), nil
}

func getConfigMapDetail(rawConfigMap *v1.ConfigMap) *ConfigMapDetail {
	return &ConfigMapDetail{
		ConfigMap:  toConfigMap(rawConfigMap.ObjectMeta),
		Data:       rawConfigMap.Data,
		BinaryData: getBinaryDataKeySize(rawConfigMap.BinaryData),
	}
}

func getBinaryDataKeySize(binaryData map[string][]byte) map[string]int {
	converted := make(map[string]int)

	for key, value := range binaryData {
		converted[key] = len(string(value))
	}

	return converted
}
