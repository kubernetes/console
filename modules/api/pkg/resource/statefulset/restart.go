package statefulset

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"time"
)

const (
	// RestartedAtAnnotationKey is an annotation key for rollout restart
	RestartedAtAnnotationKey = "kubectl.kubernetes.io/restartedAt"
)

// RestartStatefulSet restarts a daemon set in the manner of `kubectl rollout restart`.
func RestartStatefulSet(client client.Interface, namespace, name string) (*v1.StatefulSet, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if statefulSet.Spec.Template.ObjectMeta.Annotations == nil {
		statefulSet.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}
	statefulSet.Spec.Template.ObjectMeta.Annotations[RestartedAtAnnotationKey] = time.Now().Format(time.RFC3339)
	return client.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metaV1.UpdateOptions{})
}
