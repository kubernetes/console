// Copyright 2016 Google Inc. All Rights Reserved.
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

package horizontalpodautoscalerlist

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"
)

type HorizontalPodAutoscalerList struct {
	ListMeta common.ListMeta `json:"listMeta"`

	// Unordered list of Horizontal Pod Autoscalers.
	HorizontalPodAutoscalers []HorizontalPodAutoscaler `json:"horizontalpodautoscalers"`
}

// HorizontalPodAutoscaler (aka. Horizontal Pod Autoscaler)
type HorizontalPodAutoscaler struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	ScaleTargetRef ScaleTargetRef `json:"scaleTargetRef"`

	MinReplicas *int32 `json:"minReplicas"`
	MaxReplicas int32 `json:"maxReplicas"`

	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage"`
}

type ScaleTargetRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func GetHorizontalPodAutoscalerList(client k8sClient.Interface, nsQuery *common.NamespaceQuery) (*HorizontalPodAutoscalerList, error) {

   channel := common.GetHorizontalPodAutoscalerListChannel(client.Autoscaling(), nsQuery, 1)
	hpaList := <-channel.List
	if err := <-channel.Error; err != nil {
		return nil, err
	}

	return CreateHorizontalPodAutoscalerList(hpaList.Items), nil
}

func GetHorizontalPodAutoscalerListForResource(client k8sClient.Interface, namespace, kind, name string) (*HorizontalPodAutoscalerList, error) {

	nsQuery := common.NewSameNamespaceQuery(namespace)

   channel := common.GetHorizontalPodAutoscalerListChannel(client.Autoscaling(), nsQuery, 1)
	hpaList := <-channel.List
	if err := <-channel.Error; err != nil {
		return nil, err
	}

	filteredHpaList := make([]autoscaling.HorizontalPodAutoscaler, 0)
	for _, hpa := range hpaList.Items {
		if hpa.Spec.ScaleTargetRef.Kind == kind && hpa.Spec.ScaleTargetRef.Name == name {
			filteredHpaList = append(filteredHpaList, hpa)
		}
	}

	return CreateHorizontalPodAutoscalerList(filteredHpaList), nil
}

func CreateHorizontalPodAutoscalerList(hpas []autoscaling.HorizontalPodAutoscaler) *HorizontalPodAutoscalerList {
	hpaList := &HorizontalPodAutoscalerList{
		HorizontalPodAutoscalers: make([]HorizontalPodAutoscaler, 0),
		ListMeta:                 common.ListMeta{TotalItems: len(hpas)},
	}

	for _, hpa := range hpas {
		horizontalPodAutoscaler := ToHorizontalPodAutoScaler(&hpa)
		hpaList.HorizontalPodAutoscalers = append(hpaList.HorizontalPodAutoscalers, horizontalPodAutoscaler)
	}
	return hpaList
}

func ToHorizontalPodAutoScaler(hpa *autoscaling.HorizontalPodAutoscaler) HorizontalPodAutoscaler {
	return HorizontalPodAutoscaler{
		ObjectMeta: common.NewObjectMeta(hpa.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindHorizontalPodAutoscaler),

		ScaleTargetRef: ScaleTargetRef{
			Kind: hpa.Spec.ScaleTargetRef.Kind,
			Name: hpa.Spec.ScaleTargetRef.Name,
		},

		MinReplicas: hpa.Spec.MinReplicas,
		MaxReplicas: hpa.Spec.MaxReplicas,
		TargetCPUUtilizationPercentage: hpa.Spec.TargetCPUUtilizationPercentage,
	}

}
