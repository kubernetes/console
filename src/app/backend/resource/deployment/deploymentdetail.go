package deployment

import (
	"log"

	"github.com/kubernetes/dashboard/resource/common"
	"github.com/kubernetes/dashboard/resource/replicaset"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	deploymentutil "k8s.io/kubernetes/pkg/util/deployment"
)

type RollingUpdateStrategy struct {
	MaxSurge       int `json:"maxSurge"`
	MaxUnavailable int `json:"maxUnavailable"`
}

type StatusInfo struct {
	// Total number of desired replicas on the deployment
	Replicas int `json:"replicas"`

	// Number of non-terminated pods that have the desired template spec
	Updated int `json:"updated"`

	// Number of available pods (ready for at least minReadySeconds)
	// targeted by this deployment
	Available int `json:"available"`

	// Total number of unavailable pods targeted by this deployment.
	Unavailable int `json:"unavailable"`
}

// ReplicaSetDetail is a presentation layer view of Kubernetes Replica Set resource. This means
type DeploymentDetail struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`

	// Label selector of the service.
	Selector map[string]string `json:"selector"`

	// Status information on the deployment
	StatusInfo `json:"status"`

	// The deployment strategy to use to replace existing pods with new ones.
	// Valid options: Recreate, RollingUpdate
	Strategy string `json:"strategy"`

	// Min ready seconds
	MinReadySeconds int `json:"minReadySeconds"`

	// Rolling update strategy containing maxSurge and maxUnavailable
	RollingUpdateStrategy `json:"rollingUpdateStrategy,omitempty"`

	// RepliaSetList containing old replica sets from the deployment
	OldReplicaSetList replicaset.ReplicaSetList `json:"oldReplicaSetList"`

	// New replica set used by this deployment
	NewReplicaSet replicaset.ReplicaSet `json:"newReplicaSet"`

	// List of events related to this Deployment
	EventList common.EventList `json:"eventList"`
}

func GetDeploymentDetail(client client.Interface, namespace string,
	name string) (*DeploymentDetail, error) {

	log.Printf("Getting details of %s deployment in %s namespace", name, namespace)

	deploymentData, err := client.Extensions().Deployments(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	channels := &common.ResourceChannels{
		ReplicaSetList: common.GetReplicaSetListChannel(client.Extensions(), 1),
		PodList:        common.GetPodListChannel(client, 1),
	}

	replicaSetList := <-channels.ReplicaSetList.List
	if err := <-channels.ReplicaSetList.Error; err != nil {
		return nil, err
	}

	pods := <-channels.PodList.List
	if err := <-channels.PodList.Error; err != nil {
		return nil, err
	}

	oldReplicaSets, _, err := deploymentutil.FindOldReplicaSets(
		deploymentData, replicaSetList.Items, pods)
	if err != nil {
		return nil, err
	}

	newReplicaSet, err := deploymentutil.FindNewReplicaSet(deploymentData, replicaSetList.Items)
	if err != nil {
		return nil, err
	}

	events, err := GetDeploymentEvents(client, namespace, name)
	if err != nil {
		return nil, err
	}

	return getDeploymentDetail(deploymentData, oldReplicaSets, newReplicaSet,
		pods.Items, events), nil
}

func getDeploymentDetail(deployment *extensions.Deployment,
	oldRs []*extensions.ReplicaSet, newRs *extensions.ReplicaSet,
	pods []api.Pod, events *common.EventList) *DeploymentDetail {

	newRsPodInfo := common.GetPodInfo(newRs.Status.Replicas, newRs.Spec.Replicas, pods)
	newReplicaSet := toReplicaSet(newRs, &newRsPodInfo)

	oldReplicaSets := make([]extensions.ReplicaSet, len(oldRs))
	for i, replicaSet := range oldRs {
		oldReplicaSets[i] = *replicaSet
	}
	oldReplicaSetList := toReplicaSetList(oldReplicaSets, pods)

	return &DeploymentDetail{
		ObjectMeta:      common.NewObjectMeta(deployment.ObjectMeta),
		TypeMeta:        common.NewTypeMeta(common.ResourceKindDeployment),
		Selector:        deployment.Spec.Selector.MatchLabels,
		StatusInfo:      GetStatusInfo(&deployment.Status),
		Strategy:        string(deployment.Spec.Strategy.Type),
		MinReadySeconds: deployment.Spec.MinReadySeconds,
		RollingUpdateStrategy: RollingUpdateStrategy{
			MaxSurge:       deployment.Spec.Strategy.RollingUpdate.MaxSurge.IntValue(),
			MaxUnavailable: deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue(),
		},
		OldReplicaSetList: oldReplicaSetList,
		NewReplicaSet:     newReplicaSet,
		EventList:         *events,
	}
}

func toReplicaSetList(resourceList []extensions.ReplicaSet, pods []api.Pod) replicaset.ReplicaSetList {
	replicaSetList := replicaset.ReplicaSetList{
		ReplicaSets: make([]replicaset.ReplicaSet, 0),
	}

	for _, replicaSet := range resourceList {
		matchingPods := common.FilterNamespacedPodsBySelector(pods, replicaSet.ObjectMeta.Namespace,
			replicaSet.Spec.Selector.MatchLabels)
		podInfo := common.GetPodInfo(replicaSet.Status.Replicas, replicaSet.Spec.Replicas, matchingPods)

		replicaSetList.ReplicaSets = append(replicaSetList.ReplicaSets, toReplicaSet(&replicaSet, &podInfo))
	}

	return replicaSetList
}

func toReplicaSet(replicaSet *extensions.ReplicaSet, podInfo *common.PodInfo) replicaset.ReplicaSet {
	return replicaset.ReplicaSet{
		ObjectMeta:      common.NewObjectMeta(replicaSet.ObjectMeta),
		TypeMeta:        common.NewTypeMeta(common.ResourceKindReplicaSet),
		ContainerImages: common.GetContainerImages(&replicaSet.Spec.Template.Spec),
		Pods:            *podInfo,
	}
}

func GetStatusInfo(deploymentStatus *extensions.DeploymentStatus) StatusInfo {
	return StatusInfo{
		Replicas:    deploymentStatus.Replicas,
		Updated:     deploymentStatus.UpdatedReplicas,
		Available:   deploymentStatus.AvailableReplicas,
		Unavailable: deploymentStatus.UnavailableReplicas,
	}
}
