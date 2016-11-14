package deployment

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
	"k8s.io/kubernetes/pkg/util/intstr"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/horizontalpodautoscaler/horizontalpodautoscalerlist"
	"github.com/kubernetes/dashboard/src/app/backend/resource/metric"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	"github.com/kubernetes/dashboard/src/app/backend/resource/replicaset"
	"github.com/kubernetes/dashboard/src/app/backend/resource/replicaset/replicasetlist"
)

func TestGetDeploymentDetail(t *testing.T) {
	podList := &api.PodList{}
	eventList := &api.EventList{}

	deployment := &extensions.Deployment{
		ObjectMeta: api.ObjectMeta{
			Name:   "test-name",
			Labels: map[string]string{"track": "beta"},
		},
		Spec: extensions.DeploymentSpec{
			Selector:        &unversioned.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
			Replicas:        4,
			MinReadySeconds: 5,
			Strategy: extensions.DeploymentStrategy{
				Type: extensions.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &extensions.RollingUpdateDeployment{
					MaxSurge:       intstr.FromInt(1),
					MaxUnavailable: intstr.FromString("1"),
				},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   "test-pod-name",
					Labels: map[string]string{"track": "beta"},
				},
			},
		},
		Status: extensions.DeploymentStatus{
			Replicas:            4,
			UpdatedReplicas:     2,
			AvailableReplicas:   3,
			UnavailableReplicas: 1,
		},
	}

	podTemplateSpec := deploymentutil.GetNewReplicaSetTemplate(deployment)

	newReplicaSet := extensions.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name:      "replica-set-1",
			Namespace: "test-namespace",
		},
		Spec: extensions.ReplicaSetSpec{
			Template: podTemplateSpec,
		},
	}

	replicaSetList := &extensions.ReplicaSetList{
		Items: []extensions.ReplicaSet{
			newReplicaSet,
			{
				ObjectMeta: api.ObjectMeta{
					Name:      "replica-set-2",
					Namespace: "test-namespace",
				},
			},
		},
	}
	hpaList := &autoscaling.HorizontalPodAutoscalerList{
		Items: []autoscaling.HorizontalPodAutoscaler{},
	}

	cases := []struct {
		namespace, name string
		expectedActions []string
		deployment      *extensions.Deployment
		expected        *DeploymentDetail
	}{
		{
			"test-namespace", "test-name",
			[]string{"get", "list", "list", "get", "list", "get", "list", "list", "get", "list", "list", "list"},
			deployment,
			&DeploymentDetail{
				ObjectMeta: common.ObjectMeta{
					Name:   "test-name",
					Labels: map[string]string{"track": "beta"},
				},
				TypeMeta: common.TypeMeta{Kind: common.ResourceKindDeployment},
				PodList: pod.PodList{
					Pods:              []pod.Pod{},
					CumulativeMetrics: make([]metric.Metric, 0),
				},
				Selector: map[string]string{"foo": "bar"},
				StatusInfo: StatusInfo{
					Replicas:    4,
					Updated:     2,
					Available:   3,
					Unavailable: 1,
				},
				Strategy:        "RollingUpdate",
				MinReadySeconds: 5,
				RollingUpdateStrategy: &RollingUpdateStrategy{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
				OldReplicaSetList: replicasetlist.ReplicaSetList{
					ReplicaSets:       []replicaset.ReplicaSet{},
					CumulativeMetrics: make([]metric.Metric, 0),
				},
				NewReplicaSet: replicaset.ReplicaSet{
					ObjectMeta: common.NewObjectMeta(newReplicaSet.ObjectMeta),
					TypeMeta:   common.NewTypeMeta(common.ResourceKindReplicaSet),
					Pods:       common.PodInfo{Warnings: []common.Event{}},
				},
				EventList: common.EventList{
					Events: []common.Event{},
				},
				HorizontalPodAutoscalerList: horizontalpodautoscalerlist.HorizontalPodAutoscalerList{
					HorizontalPodAutoscalers: []horizontalpodautoscalerlist.HorizontalPodAutoscaler{ },
				},
			},
		},
	}

	for _, c := range cases {

		fakeClient := testclient.NewSimpleFake(c.deployment, replicaSetList, podList, eventList, hpaList)

		dataselect.DefaultDataSelectWithMetrics.MetricQuery = dataselect.NoMetrics
		actual, _ := GetDeploymentDetail(fakeClient, nil, c.namespace, c.name)

		actions := fakeClient.Actions()
		if len(actions) != len(c.expectedActions) {
			t.Errorf("Unexpected actions: %v, expected %d actions got %d", actions,
				len(c.expectedActions), len(actions))
			continue
		}

		for i, verb := range c.expectedActions {
			if actions[i].GetVerb() != verb {
				t.Errorf("Unexpected '%i' action: %+v, expected %s",
					i, actions[i], verb)
			}
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("GetDeploymentDetail(client, namespace, name) == \ngot: %#v, \nexpected %#v",
				actual, c.expected)
		}
	}
}
