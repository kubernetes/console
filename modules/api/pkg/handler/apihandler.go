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

package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/api/pkg/resource/networkpolicy"
	"k8s.io/dashboard/client"
	"k8s.io/dashboard/csrf"
	"k8s.io/dashboard/errors"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/net/xsrftoken"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/dashboard/api/pkg/handler/parser"
	"k8s.io/dashboard/api/pkg/integration"
	"k8s.io/dashboard/api/pkg/resource/clusterrole"
	"k8s.io/dashboard/api/pkg/resource/clusterrolebinding"
	"k8s.io/dashboard/api/pkg/resource/common"
	"k8s.io/dashboard/api/pkg/resource/configmap"
	"k8s.io/dashboard/api/pkg/resource/container"
	"k8s.io/dashboard/api/pkg/resource/controller"
	"k8s.io/dashboard/api/pkg/resource/cronjob"
	"k8s.io/dashboard/api/pkg/resource/customresourcedefinition"
	"k8s.io/dashboard/api/pkg/resource/customresourcedefinition/types"
	"k8s.io/dashboard/api/pkg/resource/daemonset"
	"k8s.io/dashboard/api/pkg/resource/dataselect"
	"k8s.io/dashboard/api/pkg/resource/deployment"
	"k8s.io/dashboard/api/pkg/resource/event"
	"k8s.io/dashboard/api/pkg/resource/horizontalpodautoscaler"
	"k8s.io/dashboard/api/pkg/resource/ingress"
	"k8s.io/dashboard/api/pkg/resource/ingressclass"
	"k8s.io/dashboard/api/pkg/resource/job"
	"k8s.io/dashboard/api/pkg/resource/logs"
	ns "k8s.io/dashboard/api/pkg/resource/namespace"
	"k8s.io/dashboard/api/pkg/resource/node"
	"k8s.io/dashboard/api/pkg/resource/persistentvolume"
	"k8s.io/dashboard/api/pkg/resource/persistentvolumeclaim"
	"k8s.io/dashboard/api/pkg/resource/pod"
	"k8s.io/dashboard/api/pkg/resource/replicaset"
	"k8s.io/dashboard/api/pkg/resource/replicationcontroller"
	"k8s.io/dashboard/api/pkg/resource/role"
	"k8s.io/dashboard/api/pkg/resource/rolebinding"
	"k8s.io/dashboard/api/pkg/resource/secret"
	"k8s.io/dashboard/api/pkg/resource/service"
	resourceService "k8s.io/dashboard/api/pkg/resource/service"
	"k8s.io/dashboard/api/pkg/resource/serviceaccount"
	"k8s.io/dashboard/api/pkg/resource/statefulset"
	"k8s.io/dashboard/api/pkg/resource/storageclass"
	"k8s.io/dashboard/api/pkg/scaling"
	"k8s.io/dashboard/api/pkg/validation"
)

const (
	// RequestLogString is a template for request log message.
	RequestLogString = "Incoming %s %s %s request from %s: %s"

	// ResponseLogString is a template for response log message.
	ResponseLogString = "Outgoing response to %s with %d status code"
)

// APIHandler is a representation of API handler. Structure contains clientapi and clientapi configuration.
type APIHandler struct {
	iManager integration.Manager
}

// TerminalResponse is sent by handleExecShell. The ID is a random session id that binds the original REST request and the SockJS connection.
// Any client api in possession of this ID can hijack the terminal session.
type TerminalResponse struct {
	ID string `json:"id"`
}

type JSON string

// CreateHTTPAPIHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHTTPAPIHandler(iManager integration.Manager) (*restful.Container, error) {
	apiHandler := APIHandler{iManager: iManager}
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	apiV1Ws := new(restful.WebService)

	InstallFilters(apiV1Ws)

	apiV1Ws.Path("/api/v1").
		// docs
		Doc("API v1 container").
		Param(apiV1Ws.QueryParameter("filterBy", "Comma delimited string used to apply filtering: 'propertyName,filterValue'")).
		Param(apiV1Ws.QueryParameter("sortBy", "Name of the column to sort by")).
		Param(apiV1Ws.QueryParameter("itemsPerPage", "Number of items to return when pagination is applied")).
		Param(apiV1Ws.QueryParameter("page", "Page number to return items from")).
		Param(apiV1Ws.QueryParameter("metricNames", "Metric names to download")).
		Param(apiV1Ws.QueryParameter("aggregations", "Aggregations to be performed for each metric (default: sum)")).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)

	integrationHandler := integration.NewHandler(iManager)
	integrationHandler.Install(apiV1Ws)

	// CSRF protection
	apiV1Ws.Route(
		apiV1Ws.GET("csrftoken/{action}").To(apiHandler.handleGetCsrfToken).
			// docs
			Doc("generates a one-time CSRF token that can be used by POST request").
			Param(apiV1Ws.PathParameter("action", "action name to generate CSRF token for")).
			Writes(csrf.Response{}).
			Returns(http.StatusOK, "OK", csrf.Response{}))

	// App deployment
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment").To(apiHandler.handleDeploy).
			// docs
			Doc("creates an application based on provided deployment.AppDeploymentSpec").
			Reads(deployment.AppDeploymentSpec{}).
			Writes(deployment.AppDeploymentSpec{}).
			Returns(http.StatusOK, "OK", deployment.AppDeploymentSpec{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/name").To(apiHandler.handleNameValidity).
			// docs
			Doc("checks if provided name is valid").
			Reads(validation.AppNameValiditySpec{}).
			Writes(validation.AppNameValidity{}).
			Returns(http.StatusOK, "OK", validation.AppNameValidity{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/imagereference").To(apiHandler.handleImageReferenceValidity).
			// docs
			Doc("checks if provided image is valid").
			Reads(validation.ImageReferenceValiditySpec{}).
			Writes(validation.ImageReferenceValidity{}).
			Returns(http.StatusOK, "OK", validation.ImageReferenceValidity{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeployment/validate/protocol").To(apiHandler.handleProtocolValidity).
			// docs
			Doc("checks if provided service protocol is valid").
			Reads(validation.ProtocolValiditySpec{}).
			Writes(validation.ProtocolValidity{}).
			Returns(http.StatusOK, "OK", validation.ProtocolValidity{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/appdeployment/protocols").To(apiHandler.handleGetAvailableProtocols).
			// docs
			Doc("returns a list of available protocols for the service").
			Writes(deployment.Protocols{}).
			Returns(http.StatusOK, "OK", deployment.Protocols{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/appdeploymentfromfile").To(apiHandler.handleDeployFromFile).
			// docs
			Doc("create an application from file").
			Reads(deployment.AppDeploymentFromFileSpec{}).
			Writes(deployment.AppDeploymentFromFileResponse{}).
			Returns(http.StatusOK, "OK", deployment.AppDeploymentFromFileResponse{}))

	// ReplicationController
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller").To(apiHandler.handleGetReplicationControllerList).
			// docs
			Doc("returns a list of ReplicationControllers from all namespaces").
			Writes(replicationcontroller.ReplicationControllerList{}).
			Returns(http.StatusOK, "OK", replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}").To(apiHandler.handleGetReplicationControllerList).
			// docs
			Doc("returns a list of ReplicationController in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace to get a list of ReplicationController from")).
			Writes(replicationcontroller.ReplicationControllerList{}).
			Returns(http.StatusOK, "OK", replicationcontroller.ReplicationControllerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}").To(apiHandler.handleGetReplicationControllerDetail).
			// docs
			Doc("returns detailed information about ReplicationController").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicationController")).
			Param(apiV1Ws.PathParameter("replicationController", "name of the ReplicationController")).
			Writes(replicationcontroller.ReplicationControllerDetail{}).
			Returns(http.StatusOK, "OK", replicationcontroller.ReplicationControllerDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/replicationcontroller/{namespace}/{replicationController}/update/pod").To(apiHandler.handleUpdateReplicasCount).
			// docs
			Doc("scales ReplicationController to a number of replicas").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicationController")).
			Param(apiV1Ws.PathParameter("replicationController", "name of the ReplicationController")).
			Reads(replicationcontroller.ReplicationControllerSpec{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/pod").To(apiHandler.handleGetReplicationControllerPods).
			// docs
			Doc("returns a list of Pods for ReplicationController").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicationController")).
			Param(apiV1Ws.PathParameter("replicationController", "name of the ReplicationController")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/event").To(apiHandler.handleGetReplicationControllerEvents).
			// docs
			Doc("returns a list of Events for ReplicationController").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicationController")).
			Param(apiV1Ws.PathParameter("replicationController", "name of the ReplicationController")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicationcontroller/{namespace}/{replicationController}/service").To(apiHandler.handleGetReplicationControllerServices).
			// docs
			Doc("returns a list of Services for ReplicationController").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicationController")).
			Param(apiV1Ws.PathParameter("replicationController", "name of the ReplicationController")).
			Writes(resourceService.ServiceList{}).
			Returns(http.StatusOK, "OK", resourceService.ServiceList{}))

	// ReplicaSet
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset").To(apiHandler.handleGetReplicaSets).
			// docs
			Doc("returns a list of ReplicaSets from all namespaces").
			Writes(replicaset.ReplicaSetList{}).
			Returns(http.StatusOK, "OK", replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}").To(apiHandler.handleGetReplicaSets).
			// docs
			Doc("returns a list of ReplicaSets in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicaSets")).
			Writes(replicaset.ReplicaSetList{}).
			Returns(http.StatusOK, "OK", replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}").To(apiHandler.handleGetReplicaSetDetail).
			// docs
			Doc("returns detailed information about ReplicaSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicaSet")).
			Param(apiV1Ws.PathParameter("replicaSet", "name of the ReplicaSets")).
			Writes(replicaset.ReplicaSetDetail{}).
			Returns(http.StatusOK, "OK", replicaset.ReplicaSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/pod").To(apiHandler.handleGetReplicaSetPods).
			// docs
			Doc("returns a list of Pods for ReplicaSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicaSet")).
			Param(apiV1Ws.PathParameter("replicaSet", "name of the ReplicaSets")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/service").To(apiHandler.handleGetReplicaSetServices).
			// docs
			Doc("returns a list of Services for ReplicaSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicaSet")).
			Param(apiV1Ws.PathParameter("replicaSet", "name of the ReplicaSets")).
			Writes(service.ServiceList{}).
			Returns(http.StatusOK, "OK", service.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/replicaset/{namespace}/{replicaSet}/event").To(apiHandler.handleGetReplicaSetEvents).
			// docs
			Doc("returns a list of Events for ReplicaSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ReplicaSet")).
			Param(apiV1Ws.PathParameter("replicaSet", "name of the ReplicaSets")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// Pod
	apiV1Ws.Route(
		apiV1Ws.GET("/pod").To(apiHandler.handleGetPods).
			// docs
			Doc("returns a list of Pods from all namespaces").
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}").To(apiHandler.handleGetPods).
			// docs
			Doc("returns a list of Pods in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}").To(apiHandler.handleGetPodDetail).
			// docs
			Doc("returns detailed information about Pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Writes(pod.PodDetail{}).
			Returns(http.StatusOK, "OK", pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/container").To(apiHandler.handleGetPodContainers).
			// docs
			Doc("returns a list of containers for Pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Writes(pod.PodDetail{}).
			Returns(http.StatusOK, "OK", pod.PodDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/event").To(apiHandler.handleGetPodEvents).
			// docs
			Doc("returns a list of Events for Pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/shell/{container}").To(apiHandler.handleExecShell).
			// docs
			Doc("handles exec into pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Param(apiV1Ws.PathParameter("container", "name of container in the Pod")).
			Writes(TerminalResponse{}).
			Returns(http.StatusOK, "OK", TerminalResponse{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/persistentvolumeclaim").To(apiHandler.handleGetPodPersistentVolumeClaims).
			// docs
			Doc("returns a list of containers for Pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}).
			Returns(http.StatusOK, "OK", persistentvolumeclaim.PersistentVolumeClaimList{}))

	// Deployment
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment").To(apiHandler.handleGetDeployments).
			// docs
			Doc("returns a list of Deployments from all namespaces").
			Writes(deployment.DeploymentList{}).
			Returns(http.StatusOK, "OK", deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}").To(apiHandler.handleGetDeployments).
			// docs
			Doc("returns a list of Deployments in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Writes(deployment.DeploymentList{}).
			Returns(http.StatusOK, "OK", deployment.DeploymentList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}").To(apiHandler.handleGetDeploymentDetail).
			// docs
			Doc("returns detailed information about Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(deployment.DeploymentDetail{}).
			Returns(http.StatusOK, "OK", deployment.DeploymentDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/event").To(apiHandler.handleGetDeploymentEvents).
			// docs
			Doc("returns a list of Events for Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/oldreplicaset").To(apiHandler.handleGetDeploymentOldReplicaSets).
			// docs
			Doc("returns a list of old ReplicaSets for Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(replicaset.ReplicaSetList{}).
			Returns(http.StatusOK, "OK", replicaset.ReplicaSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/deployment/{namespace}/{deployment}/newreplicaset").To(apiHandler.handleGetDeploymentNewReplicaSet).
			// docs
			Doc("returns a list of new ReplicaSets for Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(replicaset.ReplicaSet{}).
			Returns(http.StatusOK, "OK", replicaset.ReplicaSet{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/deployment/{namespace}/{deployment}/pause").To(apiHandler.handleDeploymentPause).
			// docs
			Doc("pauses the Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(deployment.DeploymentDetail{}).
			Returns(http.StatusOK, "OK", deployment.DeploymentDetail{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/deployment/{namespace}/{deployment}/rollback").To(apiHandler.handleDeploymentRollback).
			// docs
			Doc("rolls back the Deployment to the target revision").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Reads(deployment.RolloutSpec{}).
			Writes(deployment.RolloutSpec{}).
			Returns(http.StatusOK, "OK", deployment.RolloutSpec{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/deployment/{namespace}/{deployment}/restart").To(apiHandler.handleDeploymentRestart).
			// docs
			Doc("rollout restart of the Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(deployment.RolloutSpec{}).
			Returns(http.StatusOK, "OK", deployment.RolloutSpec{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/deployment/{namespace}/{deployment}/resume").To(apiHandler.handleDeploymentResume).
			// docs
			Doc("resumes the Deployment").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Deployment")).
			Param(apiV1Ws.PathParameter("deployment", "name of the Deployment")).
			Writes(deployment.DeploymentDetail{}).
			Returns(http.StatusOK, "OK", deployment.DeploymentDetail{}))

	// DaemonSet
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset").To(apiHandler.handleGetDaemonSetList).
			// docs
			Doc("returns a list of DaemonSets from all namespaces").
			Writes(daemonset.DaemonSetList{}).
			Returns(http.StatusOK, "OK", daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}").To(apiHandler.handleGetDaemonSetList).
			// docs
			Doc("returns a list of DaemonSets in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the DaemonSet")).
			Writes(daemonset.DaemonSetList{}).
			Returns(http.StatusOK, "OK", daemonset.DaemonSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}").To(apiHandler.handleGetDaemonSetDetail).
			// docs
			Doc("returns detailed information about DaemonSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the DaemonSet")).
			Param(apiV1Ws.PathParameter("daemonSet", "name of the DaemonSet")).
			Writes(daemonset.DaemonSetDetail{}).
			Returns(http.StatusOK, "OK", daemonset.DaemonSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/pod").To(apiHandler.handleGetDaemonSetPods).
			// docs
			Doc("returns a list of Pods for DaemonSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the DaemonSet")).
			Param(apiV1Ws.PathParameter("daemonSet", "name of the DaemonSet")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/service").To(apiHandler.handleGetDaemonSetServices).
			// docs
			Doc("returns a list of Services for DaemonSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the DaemonSet")).
			Param(apiV1Ws.PathParameter("daemonSet", "name of the DaemonSet")).
			Writes(resourceService.ServiceList{}).
			Returns(http.StatusOK, "OK", resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/daemonset/{namespace}/{daemonSet}/event").To(apiHandler.handleGetDaemonSetEvents).
			// docs
			Doc("returns a list of Events for DaemonSet").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the DaemonSet")).
			Param(apiV1Ws.PathParameter("daemonSet", "name of the DaemonSet")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/daemonset/{namespace}/{daemonSet}/restart").To(apiHandler.handleDaemonSetRestart).
			// docs
			Doc("rollout restart of the Daemon Set").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Daemon Set")).
			Param(apiV1Ws.PathParameter("daemonSet", "name of the Daemon Set")).
			Writes(deployment.RolloutSpec{}).
			Returns(http.StatusOK, "OK", daemonset.DaemonSetDetail{}),
	)

	// HorizontalPodAutoscaler
	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler").To(apiHandler.handleGetHorizontalPodAutoscalerList).
			// docs
			Doc("returns a list of HorizontalPodAutoscalers from all namespaces").
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerList{}).
			Returns(http.StatusOK, "OK", horizontalpodautoscaler.HorizontalPodAutoscalerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler/{namespace}").To(apiHandler.handleGetHorizontalPodAutoscalerList).
			// docs
			Doc("returns a list of HorizontalPodAutoscalers in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the HorizontalPodAutoscaler")).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerList{}).
			Returns(http.StatusOK, "OK", horizontalpodautoscaler.HorizontalPodAutoscalerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/{kind}/{namespace}/{name}/horizontalpodautoscaler").To(apiHandler.handleGetHorizontalPodAutoscalerListForResource).
			// docs
			Doc("returns a list of HorizontalPodAutoscalers for resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource to get HorizontalPodAutoscalers for")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource to get HorizontalPodAutoscalers for")).
			Param(apiV1Ws.PathParameter("name", "name of the resource to get HorizontalPodAutoscalers for")).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerList{}).
			Returns(http.StatusOK, "OK", horizontalpodautoscaler.HorizontalPodAutoscalerList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/horizontalpodautoscaler/{namespace}/{horizontalpodautoscaler}").To(apiHandler.handleGetHorizontalPodAutoscalerDetail).
			// docs
			Doc("returns detailed information about HorizontalPodAutoscaler").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the HorizontalPodAutoscaler")).
			Param(apiV1Ws.PathParameter("horizontalpodautoscaler", "name of the HorizontalPodAutoscaler")).
			Writes(horizontalpodautoscaler.HorizontalPodAutoscalerDetail{}).
			Returns(http.StatusOK, "OK", horizontalpodautoscaler.HorizontalPodAutoscalerDetail{}))

	// Job
	apiV1Ws.Route(
		apiV1Ws.GET("/job").To(apiHandler.handleGetJobList).
			// docs
			Doc("returns a list of Jobs from all namespaces").
			Writes(job.JobList{}).
			Returns(http.StatusOK, "OK", job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}").To(apiHandler.handleGetJobList).
			// docs
			Doc("returns a list of Jobs in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Job")).
			Writes(job.JobList{}).
			Returns(http.StatusOK, "OK", job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}").To(apiHandler.handleGetJobDetail).
			// docs
			Doc("returns detailed information about Job").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Job")).
			Param(apiV1Ws.PathParameter("name", "name of the Job")).
			Writes(job.JobDetail{}).
			Returns(http.StatusOK, "OK", job.JobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}/pod").To(apiHandler.handleGetJobPods).
			// docs
			Doc("returns a list of Pods for Job").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Job")).
			Param(apiV1Ws.PathParameter("name", "name of the Job")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/job/{namespace}/{name}/event").To(apiHandler.handleGetJobEvents).
			// docs
			Doc("returns a list of Events for Job").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Job")).
			Param(apiV1Ws.PathParameter("name", "name of the Job")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// CronJob
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob").To(apiHandler.handleGetCronJobList).
			// docs
			Doc("returns a list of CronJobs from all namespaces").
			Writes(cronjob.CronJobList{}).
			Returns(http.StatusOK, "OK", cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}").To(apiHandler.handleGetCronJobList).
			// docs
			Doc("returns a list of CronJobs in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the CronJob")).
			Writes(cronjob.CronJobList{}).
			Returns(http.StatusOK, "OK", cronjob.CronJobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}").To(apiHandler.handleGetCronJobDetail).
			// docs
			Doc("returns detailed information about CronJob").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the CronJob")).
			Param(apiV1Ws.PathParameter("name", "name of the CronJob")).
			Writes(cronjob.CronJobDetail{}).
			Returns(http.StatusOK, "OK", cronjob.CronJobDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}/job").To(apiHandler.handleGetCronJobJobs).
			// docs
			Doc("returns a list of Jobs for CronJob").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the CronJob")).
			Param(apiV1Ws.PathParameter("name", "name of the CronJob")).
			Param(apiV1Ws.QueryParameter("active", "filter related Jobs by active status")).
			Writes(job.JobList{}).
			Returns(http.StatusOK, "OK", job.JobList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/cronjob/{namespace}/{name}/event").To(apiHandler.handleGetCronJobEvents).
			// docs
			Doc("returns a list of Events for CronJob").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the CronJob")).
			Param(apiV1Ws.PathParameter("name", "name of the CronJob")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/cronjob/{namespace}/{name}/trigger").To(apiHandler.handleTriggerCronJob).
			// docs
			Doc("triggers a Job based on CronJob").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the CronJob")).
			Param(apiV1Ws.PathParameter("name", "name of the CronJob")).
			Returns(http.StatusOK, "OK", nil))

	// Namespace
	apiV1Ws.Route(
		apiV1Ws.POST("/namespace").To(apiHandler.handleCreateNamespace).
			// docs
			Doc("create a Namespace").
			Reads(ns.NamespaceSpec{}).
			Writes(ns.NamespaceSpec{}).
			Returns(http.StatusOK, "OK", ns.NamespaceSpec{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace").To(apiHandler.handleGetNamespaces).
			// docs
			Doc("returns a list of Namespaces").
			Writes(ns.NamespaceList{}).
			Returns(http.StatusOK, "OK", ns.NamespaceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace/{name}").To(apiHandler.handleGetNamespaceDetail).
			// docs
			Doc("returns detailed information about Namespace").
			Param(apiV1Ws.PathParameter("name", "name of the Namespace")).
			Writes(ns.NamespaceDetail{}).
			Returns(http.StatusOK, "OK", ns.NamespaceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/namespace/{name}/event").To(apiHandler.handleGetNamespaceEvents).
			// docs
			Doc("returns a list of Events for Namespace").
			Param(apiV1Ws.PathParameter("name", "name of the Namespace")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// Event
	apiV1Ws.Route(
		apiV1Ws.GET("/event").To(apiHandler.handleGetEventList).
			// docs
			Doc("returns a list of Events from all namespaces").
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/event/{namespace}").To(apiHandler.handleGetEventList).
			// docs
			Doc("returns a list of Events in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace to get Events from")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// Secret
	apiV1Ws.Route(
		apiV1Ws.GET("/secret").To(apiHandler.handleGetSecretList).
			// docs
			Doc("returns a list of Secrets from all namespaces").
			Writes(secret.SecretList{}).
			Returns(http.StatusOK, "OK", secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/secret/{namespace}").To(apiHandler.handleGetSecretList).
			// docs
			Doc("returns a list of Secrets in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Secret")).
			Writes(secret.SecretList{}).
			Returns(http.StatusOK, "OK", secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/secret/{namespace}/{name}").To(apiHandler.handleGetSecretDetail).
			// docs
			Doc("returns detailed information about Secret").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Secret")).
			Param(apiV1Ws.PathParameter("name", "name of the Secret")).
			Writes(secret.SecretDetail{}).
			Returns(http.StatusOK, "OK", secret.SecretDetail{}))
	apiV1Ws.Route(
		apiV1Ws.POST("/secret").To(apiHandler.handleCreateImagePullSecret).
			// docs
			Doc("stores ImagePullSecret in a Kubernetes Secret").
			Reads(secret.ImagePullSecretSpec{}).
			Writes(secret.Secret{}).
			Returns(http.StatusOK, "OK", secret.Secret{}))

	// ConfigMap
	apiV1Ws.Route(
		apiV1Ws.GET("/configmap").To(apiHandler.handleGetConfigMapList).
			// docs
			Doc("returns a list of ConfigMaps from all namespaces").
			Writes(configmap.ConfigMapList{}).
			Returns(http.StatusOK, "OK", configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/configmap/{namespace}").To(apiHandler.handleGetConfigMapList).
			// docs
			Doc("returns a list of ConfigMaps in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ConfigMap")).
			Writes(configmap.ConfigMapList{}).
			Returns(http.StatusOK, "OK", configmap.ConfigMapList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/configmap/{namespace}/{configmap}").To(apiHandler.handleGetConfigMapDetail).
			// docs
			Doc("returns detailed information about ConfigMap").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ConfigMap")).
			Param(apiV1Ws.PathParameter("configmap", "name of the ConfigMap")).
			Writes(configmap.ConfigMapDetail{}).
			Returns(http.StatusOK, "OK", configmap.ConfigMapDetail{}))

	// Service
	apiV1Ws.Route(
		apiV1Ws.GET("/service").To(apiHandler.handleGetServiceList).
			// docs
			Doc("returns a list of Services from all namespaces").
			Writes(resourceService.ServiceList{}).
			Returns(http.StatusOK, "OK", resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}").To(apiHandler.handleGetServiceList).
			// docs
			Doc("returns a list of Services in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Service")).
			Writes(resourceService.ServiceList{}).
			Returns(http.StatusOK, "OK", resourceService.ServiceList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}").To(apiHandler.handleGetServiceDetail).
			// docs
			Doc("returns detailed information about Service").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Service")).
			Param(apiV1Ws.PathParameter("service", "name of the Service")).
			Writes(resourceService.ServiceDetail{}).
			Returns(http.StatusOK, "OK", resourceService.ServiceDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}/event").To(apiHandler.handleGetServiceEvent).
			// docs
			Doc("returns a list of Events for Service").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Service")).
			Param(apiV1Ws.PathParameter("service", "name of the Service")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}/pod").To(apiHandler.handleGetServicePods).
			// docs
			Doc("returns a list of Pods for Service").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Service")).
			Param(apiV1Ws.PathParameter("service", "name of the Service")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/service/{namespace}/{service}/ingress").To(apiHandler.handleGetServiceIngressList).
			// docs
			Doc("returns a list of Ingresses for Service").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Service")).
			Param(apiV1Ws.PathParameter("service", "name of the Service")).
			Writes(ingress.IngressList{}).
			Returns(http.StatusOK, "OK", ingress.IngressList{}))

	// ServiceAccount
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount").To(apiHandler.handleGetServiceAccountList).
			// docs
			Doc("returns a list of ServiceAccounts from all namespaces").
			Writes(serviceaccount.ServiceAccountList{}).
			Returns(http.StatusOK, "OK", serviceaccount.ServiceAccountList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount/{namespace}").To(apiHandler.handleGetServiceAccountList).
			// docs
			Doc("returns a list of ServiceAccounts in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ServiceAccount")).
			Writes(serviceaccount.ServiceAccountList{}).
			Returns(http.StatusOK, "OK", serviceaccount.ServiceAccountList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount/{namespace}/{serviceaccount}").To(apiHandler.handleGetServiceAccountDetail).
			// docs
			Doc("returns detailed information about ServiceAccount").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ServiceAccount")).
			Param(apiV1Ws.PathParameter("serviceaccount", "name of the ServiceAccount")).
			Writes(serviceaccount.ServiceAccountDetail{}).
			Returns(http.StatusOK, "OK", serviceaccount.ServiceAccountDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount/{namespace}/{serviceaccount}/secret").To(apiHandler.handleGetServiceAccountSecrets).
			// docs
			Doc("returns a list of Secrets for ServiceAccount").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ServiceAccount")).
			Param(apiV1Ws.PathParameter("serviceaccount", "name of the ServiceAccount")).
			Writes(secret.SecretList{}).
			Returns(http.StatusOK, "OK", secret.SecretList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/serviceaccount/{namespace}/{serviceaccount}/imagepullsecret").To(apiHandler.handleGetServiceAccountImagePullSecrets).
			// docs
			Doc("returns a list of ImagePullSecret Secrets for ServiceAccount").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the ServiceAccount")).
			Param(apiV1Ws.PathParameter("serviceaccount", "name of the ServiceAccount")).
			Writes(secret.SecretList{}).
			Returns(http.StatusOK, "OK", secret.SecretList{}))

	// Ingress
	apiV1Ws.Route(apiV1Ws.GET("/ingress").To(apiHandler.handleGetIngressList).
		// docs
		Doc("returns a list of Ingresses from all namespaces").
		Writes(ingress.IngressList{}).
		Returns(http.StatusOK, "OK", ingress.IngressList{}))
	apiV1Ws.Route(apiV1Ws.GET("/ingress/{namespace}").To(apiHandler.handleGetIngressList).
		// docs
		Doc("returns a list of Ingresses in a namespaces").
		Param(apiV1Ws.PathParameter("namespace", "namespace of the Ingress")).
		Writes(ingress.IngressList{}).
		Returns(http.StatusOK, "OK", ingress.IngressList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/ingress/{namespace}/{name}").To(apiHandler.handleGetIngressDetail).
			// docs
			Doc("returns detailed information about Ingress").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Ingress")).
			Param(apiV1Ws.PathParameter("name", "name of the Ingress")).
			Writes(ingress.IngressDetail{}).
			Returns(http.StatusOK, "OK", ingress.IngressDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/ingress/{namespace}/{ingress}/event").To(apiHandler.handleGetIngressEvent).
			// docs
			Doc("returns a list of Events for Ingress").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Ingress")).
			Param(apiV1Ws.PathParameter("ingress", "name of the Ingress")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// NetworkPolicy
	apiV1Ws.Route(
		apiV1Ws.GET("/networkpolicy").To(apiHandler.handleGetNetworkPolicyList).
			// docs
			Doc("returns a list of NetworkPolicies from all namespaces").
			Writes(networkpolicy.NetworkPolicyList{}).
			Returns(http.StatusOK, "OK", networkpolicy.NetworkPolicyList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/networkpolicy/{namespace}").To(apiHandler.handleGetNetworkPolicyList).
			// docs
			Doc("returns a list of NetworkPolicies in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the NetworkPolicy")).
			Writes(networkpolicy.NetworkPolicyList{}).
			Returns(http.StatusOK, "OK", networkpolicy.NetworkPolicyList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/networkpolicy/{namespace}/{networkpolicy}").To(apiHandler.handleGetNetworkPolicyDetail).
			// docs
			Doc("returns detailed information about NetworkPolicy").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the NetworkPolicy")).
			Param(apiV1Ws.PathParameter("networkpolicy", "name of the NetworkPolicy")).
			Writes(networkpolicy.NetworkPolicyDetail{}).
			Returns(http.StatusOK, "OK", networkpolicy.NetworkPolicyDetail{}))

	// StatefulSet
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset").To(apiHandler.handleGetStatefulSetList).
			// docs
			Doc("returns a list of StatefulSets from all namespaces").
			Writes(statefulset.StatefulSetList{}).
			Returns(http.StatusOK, "OK", statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}").To(apiHandler.handleGetStatefulSetList).
			// docs
			Doc("returns a list of StatefulSets in a namespaces").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the StatefulSet")).
			Writes(statefulset.StatefulSetList{}).
			Returns(http.StatusOK, "OK", statefulset.StatefulSetList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}").To(apiHandler.handleGetStatefulSetDetail).
			// docs
			Doc("returns detailed information about StatefulSets").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the StatefulSet")).
			Param(apiV1Ws.PathParameter("statefulset", "name of the StatefulSet")).
			Writes(statefulset.StatefulSetDetail{}).
			Returns(http.StatusOK, "OK", statefulset.StatefulSetDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}/pod").To(apiHandler.handleGetStatefulSetPods).
			// docs
			Doc("returns  a list of Pods for StatefulSets").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the StatefulSet")).
			Param(apiV1Ws.PathParameter("statefulset", "name of the StatefulSet")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/statefulset/{namespace}/{statefulset}/event").To(apiHandler.handleGetStatefulSetEvents).
			// docs
			Doc("returns a list of Events for StatefulSets").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the StatefulSet")).
			Param(apiV1Ws.PathParameter("statefulset", "name of the StatefulSet")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/statefulset/{namespace}/{statefulset}/restart").To(apiHandler.handleStatefulSetRestart).
			// docs
			Doc("rollout restart of the Daemon Set").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the StatefulSet")).
			Param(apiV1Ws.PathParameter("statefulset", "name of the StatefulSet")).
			Writes(deployment.RolloutSpec{}).
			Returns(http.StatusOK, "OK", statefulset.StatefulSetDetail{}),
	)

	// Node
	apiV1Ws.Route(
		apiV1Ws.GET("/node").To(apiHandler.handleGetNodeList).
			// docs
			Doc("returns a list of Nodes").
			Writes(node.NodeList{}).
			Returns(http.StatusOK, "OK", node.NodeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}").To(apiHandler.handleGetNodeDetail).
			// docs
			Doc("returns detailed information about Node").
			Param(apiV1Ws.PathParameter("name", "name of the Node")).
			Writes(node.NodeDetail{}).
			Returns(http.StatusOK, "OK", node.NodeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/event").To(apiHandler.handleGetNodeEvents).
			// docs
			Doc("returns a list of Events for Node").
			Param(apiV1Ws.PathParameter("name", "name of the Node")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/pod").To(apiHandler.handleGetNodePods).
			// docs
			Doc("returns a list of Pods for Node").
			Param(apiV1Ws.PathParameter("name", "name of the Node")).
			Writes(pod.PodList{}).
			Returns(http.StatusOK, "OK", pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/node/{name}/drain").To(apiHandler.handleNodeDrain).
			// docs
			Doc("drains Node").
			Param(apiV1Ws.PathParameter("name", "name of the Node")).
			Reads(node.NodeDrainSpec{}).
			Returns(http.StatusOK, "OK", nil))

	// Verber (namespaced)
	apiV1Ws.Route(
		apiV1Ws.DELETE("/_raw/{kind}/namespace/{namespace}/name/{name}").To(apiHandler.handleDeleteResource).
			// docs
			Doc("deletes a resource from a namespace").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Param(apiV1Ws.QueryParameter("deleteNow", "override graceful delete options and enforce immediate deletion")).
			Param(apiV1Ws.QueryParameter("propagation", "override default delete propagation policy")).
			Returns(http.StatusNoContent, "", nil))
	apiV1Ws.Route(
		apiV1Ws.GET("/_raw/{kind}/namespace/{namespace}/name/{name}").To(apiHandler.handleGetResource).
			// docs
			Doc("returns unstructured resource from a namespace").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Writes(unstructured.Unstructured{}).
			Returns(http.StatusOK, "OK", unstructured.Unstructured{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/_raw/{kind}/namespace/{namespace}/name/{name}").To(apiHandler.handlePutResource).
			// docs
			Doc("creates or updates a resource in a namespace").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Reads(JSON("")).
			Returns(http.StatusNoContent, "", nil))

	// Verber (non-namespaced)
	apiV1Ws.Route(
		apiV1Ws.DELETE("/_raw/{kind}/name/{name}").To(apiHandler.handleDeleteResource).
			// docs
			Doc("deletes a non-namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Param(apiV1Ws.QueryParameter("deleteNow", "override graceful delete options and enforce immediate deletion")).
			Param(apiV1Ws.QueryParameter("propagation", "override default delete propagation policy")).
			Returns(http.StatusNoContent, "", nil))
	apiV1Ws.Route(
		apiV1Ws.GET("/_raw/{kind}/name/{name}").To(apiHandler.handleGetResource).
			// docs
			Doc("returns a non-namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Writes(unstructured.Unstructured{}).
			Returns(http.StatusOK, "OK", unstructured.Unstructured{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/_raw/{kind}/name/{name}").To(apiHandler.handlePutResource).
			// docs
			Doc("creates or updates a non-namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Reads(JSON("")).
			Returns(http.StatusNoContent, "", nil))

	// Generic resource scaling
	apiV1Ws.Route(
		apiV1Ws.PUT("/scale/{kind}/{namespace}/{name}").To(apiHandler.handleScaleResource).
			// docs
			Doc("scales a namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Param(apiV1Ws.QueryParameter("scaleBy", "desired number of replicas")).
			Writes(scaling.ReplicaCounts{}).
			Returns(http.StatusOK, "OK", scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.PUT("/scale/{kind}/{name}").To(apiHandler.handleScaleResource).
			// docs
			Doc("scales a non-namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Param(apiV1Ws.QueryParameter("scaleBy", "desired number of replicas")).
			Writes(scaling.ReplicaCounts{}).
			Returns(http.StatusOK, "OK", scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/scale/{kind}/{namespace}/{name}").To(apiHandler.handleGetReplicaCount).
			// docs
			Doc("returns a number of replicas of namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Writes(scaling.ReplicaCounts{}).
			Returns(http.StatusOK, "OK", scaling.ReplicaCounts{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/scale/{kind}/{name}").To(apiHandler.handleGetReplicaCount).
			// docs
			Doc("returns a number of replicas of non-namespaced resource").
			Param(apiV1Ws.PathParameter("kind", "kind of the resource")).
			Param(apiV1Ws.PathParameter("name", "name of the resource")).
			Writes(scaling.ReplicaCounts{}).
			Returns(http.StatusOK, "OK", scaling.ReplicaCounts{}))

	// ClusterRole
	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrole").To(apiHandler.handleGetClusterRoleList).
			// docs
			Doc("returns a list of ClusterRoles").
			Writes(clusterrole.ClusterRoleList{}).
			Returns(http.StatusOK, "OK", clusterrole.ClusterRoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrole/{name}").To(apiHandler.handleGetClusterRoleDetail).
			// docs
			Doc("returns detailed information about ClusterRole").
			Param(apiV1Ws.PathParameter("name", "name of the ClusterRole")).
			Writes(clusterrole.ClusterRoleDetail{}).
			Returns(http.StatusOK, "OK", clusterrole.ClusterRoleDetail{}))

	// ClusterRoleBinding
	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrolebinding").To(apiHandler.handleGetClusterRoleBindingList).
			// docs
			Doc("returns a list of ClusterRoleBindings").
			Writes(clusterrolebinding.ClusterRoleBindingList{}).
			Returns(http.StatusOK, "OK", clusterrolebinding.ClusterRoleBindingList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/clusterrolebinding/{name}").To(apiHandler.handleGetClusterRoleBindingDetail).
			// docs
			Doc("returns detailed information about ClusterRoleBinding").
			Param(apiV1Ws.PathParameter("name", "name of the ClusterRoleBinding")).
			Writes(clusterrolebinding.ClusterRoleBindingDetail{}).
			Returns(http.StatusOK, "OK", clusterrolebinding.ClusterRoleBindingDetail{}))

	// Role
	apiV1Ws.Route(
		apiV1Ws.GET("/role").To(apiHandler.handleGetRoleList).
			// docs
			Doc("returns a list of Roles from all namespace").
			Writes(role.RoleList{}).
			Returns(http.StatusOK, "OK", role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/role/{namespace}").To(apiHandler.handleGetRoleList).
			// docs
			Doc("returns a list of Roles in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Role")).
			Writes(role.RoleList{}).
			Returns(http.StatusOK, "OK", role.RoleList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/role/{namespace}/{name}").To(apiHandler.handleGetRoleDetail).
			// docs
			Doc("returns detailed information about Role").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Role")).
			Param(apiV1Ws.PathParameter("name", "name of the Role")).
			Writes(role.RoleDetail{}).
			Returns(http.StatusOK, "OK", role.RoleDetail{}))

	// RoleBinding
	apiV1Ws.Route(
		apiV1Ws.GET("/rolebinding").To(apiHandler.handleGetRoleBindingList).
			// docs
			Doc("returns a list of RoleBindings from all namespace").
			Writes(rolebinding.RoleBindingList{}).
			Returns(http.StatusOK, "OK", rolebinding.RoleBindingList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/rolebinding/{namespace}").To(apiHandler.handleGetRoleBindingList).
			// docs
			Doc("returns a list of RoleBindings in a namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the RoleBinding")).
			Writes(rolebinding.RoleBindingList{}).
			Returns(http.StatusOK, "OK", rolebinding.RoleBindingList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/rolebinding/{namespace}/{name}").To(apiHandler.handleGetRoleBindingDetail).
			// docs
			Doc("returns detailed information about RoleBinding").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the RoleBinding")).
			Param(apiV1Ws.PathParameter("name", "name of the RoleBinding")).
			Writes(rolebinding.RoleBindingDetail{}).
			Returns(http.StatusOK, "OK", rolebinding.RoleBindingDetail{}))

	// PersistentVolume
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume").To(apiHandler.handleGetPersistentVolumeList).
			// docs
			Doc("returns a list of PersistentVolumes from all namespaces").
			Writes(persistentvolume.PersistentVolumeList{}).
			Returns(http.StatusOK, "OK", persistentvolume.PersistentVolumeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume/{persistentvolume}").To(apiHandler.handleGetPersistentVolumeDetail).
			// docs
			Doc("returns detailed information about PersistentVolume").
			Param(apiV1Ws.PathParameter("persistentvolume", "name of the PersistentVolume")).
			Writes(persistentvolume.PersistentVolumeDetail{}).
			Returns(http.StatusOK, "OK", persistentvolume.PersistentVolumeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolume/namespace/{namespace}/name/{persistentvolume}").To(apiHandler.handleGetPersistentVolumeDetail).
			// docs
			Doc("returns detailed information about PersistentVolume").
			Param(apiV1Ws.PathParameter("persistentvolume", "name of the PersistentVolume")).
			Writes(persistentvolume.PersistentVolumeDetail{}).
			Returns(http.StatusOK, "OK", persistentvolume.PersistentVolumeDetail{}))

	// PersistentVolumeClaim
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			// docs
			Doc("returns a list of PersistentVolumeClaim").
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}).
			Returns(http.StatusOK, "OK", persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/{namespace}").
			To(apiHandler.handleGetPersistentVolumeClaimList).
			// docs
			Doc("returns a list of PersistentVolumeClaim from specified namespace").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the PersistentVolumeClaim")).
			Writes(persistentvolumeclaim.PersistentVolumeClaimList{}).
			Returns(http.StatusOK, "OK", persistentvolumeclaim.PersistentVolumeClaimList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/persistentvolumeclaim/{namespace}/{name}").
			To(apiHandler.handleGetPersistentVolumeClaimDetail).
			// docs
			Doc("returns detailed information about PersistentVolumeClaim").
			Param(apiV1Ws.PathParameter("name", "name of the PersistentVolumeClaim")).
			Param(apiV1Ws.PathParameter("namespace", "namespace of the PersistentVolumeClaim")).
			Writes(persistentvolumeclaim.PersistentVolumeClaimDetail{}).
			Returns(http.StatusOK, "OK", persistentvolumeclaim.PersistentVolumeClaimDetail{}))

	// CRD
	apiV1Ws.Route(
		apiV1Ws.GET("/crd").
			To(apiHandler.handleGetCustomResourceDefinitionList).
			// docs
			Doc("returns a list of CustomResourceDefinition").
			Writes(types.CustomResourceDefinitionList{}).
			Returns(http.StatusOK, "OK", types.CustomResourceDefinitionList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{crd}").
			To(apiHandler.handleGetCustomResourceDefinitionDetail).
			// docs
			Doc("returns detailed information about CustomResourceDefinition").
			Param(apiV1Ws.PathParameter("crd", "name of the CustomResourceDefinition")).
			Writes(types.CustomResourceDefinitionDetail{}).
			Returns(http.StatusOK, "OK", types.CustomResourceDefinitionDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/object").
			To(apiHandler.handleGetCustomResourceObjectList).
			// docs
			Doc("returns a list of objects of CustomResourceDefinition").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the custom resource")).
			Param(apiV1Ws.PathParameter("crd", "name of the CustomResourceDefinition")).
			Writes(types.CustomResourceObjectList{}).
			Returns(http.StatusOK, "OK", types.CustomResourceObjectList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/{object}").
			To(apiHandler.handleGetCustomResourceObjectDetail).
			// docs
			Doc("returns detailed information about custom resource object").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the custom resource")).
			Param(apiV1Ws.PathParameter("crd", "name of the CustomResourceDefinition")).
			Param(apiV1Ws.PathParameter("object", "name of the custom resource object")).
			Writes(types.CustomResourceObjectDetail{}).
			Returns(http.StatusOK, "OK", types.CustomResourceObjectDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/crd/{namespace}/{crd}/{object}/event").
			To(apiHandler.handleGetCustomResourceObjectEvents).
			// docs
			Doc("returns Events for custom resource object").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the custom resource")).
			Param(apiV1Ws.PathParameter("crd", "name of the CustomResourceDefinition")).
			Param(apiV1Ws.PathParameter("object", "name of the custom resource object")).
			Writes(common.EventList{}).
			Returns(http.StatusOK, "OK", common.EventList{}))

	// StorageClass
	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass").
			To(apiHandler.handleGetStorageClassList).
			// docs
			Doc("returns a list of StorageClasses").
			Writes(storageclass.StorageClassList{}).
			Returns(http.StatusOK, "OK", storageclass.StorageClassList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass/{storageclass}").
			// docs
			Doc("returns detailed information about StorageClass").
			Param(apiV1Ws.PathParameter("storageclass", "name of the StorageClass")).
			To(apiHandler.handleGetStorageClass).
			Writes(storageclass.StorageClass{}).
			Returns(http.StatusOK, "OK", storageclass.StorageClass{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/storageclass/{storageclass}/persistentvolume").
			To(apiHandler.handleGetStorageClassPersistentVolumes).
			// docs
			Doc("returns a list of PersistentVolumes assigned to StorageClass").
			Param(apiV1Ws.PathParameter("storageclass", "name of the StorageClass")).
			Writes(persistentvolume.PersistentVolumeList{}).
			Returns(http.StatusOK, "OK", persistentvolume.PersistentVolumeList{}))

	// IngressClass
	apiV1Ws.Route(
		apiV1Ws.GET("/ingressclass").
			To(apiHandler.handleGetIngressClassList).
			// docs
			Doc("returns a list of IngressClasses").
			Writes(ingressclass.IngressClassList{}).
			Returns(http.StatusOK, "OK", ingressclass.IngressClassList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/ingressclass/{ingressclass}").
			To(apiHandler.handleGetIngressClass).
			// docs
			Doc("returns detailed information about IngressClass").
			Param(apiV1Ws.PathParameter("ingressclass", "name of the IngressClass")).
			Writes(ingressclass.IngressClass{}).
			Returns(http.StatusOK, "OK", ingressclass.IngressClass{}))

	// Logs
	apiV1Ws.Route(
		apiV1Ws.GET("/log/source/{namespace}/{resourceName}/{resourceType}").
			To(apiHandler.handleLogSource).
			// docs
			Doc("returns log sources for a resource").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the resource")).
			Param(apiV1Ws.PathParameter("resourceName", "name of the resource")).
			Param(apiV1Ws.PathParameter("resourceType", "type of the resource")).
			Writes(controller.LogSources{}).
			Returns(http.StatusOK, "OK", controller.LogSources{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/{namespace}/{pod}").
			To(apiHandler.handleLogs).
			// docs
			Doc("returns logs from a Pod").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Writes(logs.LogDetails{}).
			Returns(http.StatusOK, "OK", logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogs).
			// docs
			Doc("returns logs from a Container").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Param(apiV1Ws.PathParameter("container", "name of container in the Pod")).
			Writes(logs.LogDetails{}).
			Returns(http.StatusOK, "OK", logs.LogDetails{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/log/file/{namespace}/{pod}/{container}").
			To(apiHandler.handleLogFile).
			// docs
			Doc("returns a text file with logs from a Container").
			Param(apiV1Ws.PathParameter("namespace", "namespace of the Pod")).
			Param(apiV1Ws.PathParameter("pod", "name of the Pod")).
			Param(apiV1Ws.PathParameter("container", "name of container in the Pod")).
			Writes([]byte{}).
			Returns(http.StatusOK, "OK", []byte{}))

	return wsContainer, nil
}

func (apiHandler *APIHandler) handleGetClusterRoleList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := clusterrole.GetClusterRoleList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetClusterRoleDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	result, err := clusterrole.GetClusterRoleDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetClusterRoleBindingList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := clusterrolebinding.GetClusterRoleBindingList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetClusterRoleBindingDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	result, err := clusterrolebinding.GetClusterRoleBindingDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetRoleList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := role.GetRoleList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetRoleDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := role.GetRoleDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetRoleBindingList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := rolebinding.GetRoleBindingList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetRoleBindingDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := rolebinding.GetRoleBindingDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCsrfToken(request *restful.Request, response *restful.Response) {
	action := request.PathParameter("action")
	token := xsrftoken.Generate(csrf.Key(), "none", action)
	_ = response.WriteHeaderAndEntity(http.StatusOK, csrf.Response{Token: token})
}

func (apiHandler *APIHandler) handleGetStatefulSetList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetList(k8sClient, namespace, dataSelect,
		apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetStatefulSetDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	result, err := statefulset.GetStatefulSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetStatefulSetPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := statefulset.GetStatefulSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetStatefulSetEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statefulset")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := resourceService.GetServiceList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	result, err := resourceService.GetServiceDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceEvent(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServiceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceAccountList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := serviceaccount.GetServiceAccountList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceAccountDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("serviceaccount")
	result, err := serviceaccount.GetServiceAccountDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceAccountImagePullSecrets(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("serviceaccount")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := serviceaccount.GetServiceAccountImagePullSecrets(k8sClient, namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceAccountSecrets(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("serviceaccount")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := serviceaccount.GetServiceAccountSecrets(k8sClient, namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetIngressDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := ingress.GetIngressDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetIngressEvent(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("ingress")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.NoMetrics
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetIngressList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := ingress.GetIngressList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServicePods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := resourceService.GetServicePods(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetServiceIngressList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("service")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.NoMetrics
	result, err := resourceService.GetServiceIngressList(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNetworkPolicyList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := networkpolicy.GetNetworkPolicyList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNetworkPolicyDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("networkpolicy")
	result, err := networkpolicy.GetNetworkPolicyDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := node.GetNodeList(k8sClient, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := node.GetNodeDetail(k8sClient, apiHandler.iManager.Metric().Client(), name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := event.GetNodeEvents(k8sClient, dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodePods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := node.GetNodePods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleNodeDrain(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	spec := new(node.NodeDrainSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := node.DrainNode(k8sClient, name, spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func (apiHandler *APIHandler) handleDeploy(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	appDeploymentSpec := new(deployment.AppDeploymentSpec)
	if err := request.ReadEntity(appDeploymentSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := deployment.DeployApp(appDeploymentSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusCreated, appDeploymentSpec)
}

func (apiHandler *APIHandler) handleScaleResource(request *restful.Request, response *restful.Response) {
	cfg, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	count := request.QueryParameter("scaleBy")
	replicaCountSpec, err := scaling.ScaleResource(cfg, kind, namespace, name, count)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, replicaCountSpec)
}

func (apiHandler *APIHandler) handleGetReplicaCount(request *restful.Request, response *restful.Response) {
	cfg, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	replicaCounts, err := scaling.GetReplicaCounts(cfg, kind, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, replicaCounts)
}

func (apiHandler *APIHandler) handleDeployFromFile(request *restful.Request, response *restful.Response) {
	cfg, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	deploymentSpec := new(deployment.AppDeploymentFromFileSpec)
	if err := request.ReadEntity(deploymentSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	isDeployed, err := deployment.DeployAppFromFile(cfg, deploymentSpec)
	if !isDeployed {
		errors.HandleInternalError(response, err)
		return
	}

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	_ = response.WriteHeaderAndEntity(http.StatusCreated, deployment.AppDeploymentFromFileResponse{
		Name:    deploymentSpec.Name,
		Content: deploymentSpec.Content,
		Error:   errorMessage,
	})
}

func (apiHandler *APIHandler) handleDeploymentPause(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	deploymentSpec, err := deployment.PauseDeployment(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, deploymentSpec)
}

func (apiHandler *APIHandler) handleDeploymentRollback(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	rolloutSpec := new(deployment.RolloutSpec)
	if err := request.ReadEntity(rolloutSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	rolloutSpec, err = deployment.RollbackDeployment(k8sClient, rolloutSpec, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, rolloutSpec)
}

func (apiHandler *APIHandler) handleDeploymentRestart(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	rolloutSpec, err := deployment.RestartDeployment(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, rolloutSpec)
}

func (apiHandler *APIHandler) handleDeploymentResume(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	deploymentSpec, err := deployment.ResumeDeployment(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, deploymentSpec)
}

func (apiHandler *APIHandler) handleNameValidity(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	spec := new(validation.AppNameValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	validity, err := validation.ValidateAppName(spec, k8sClient)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, validity)
}

func (apiHandler *APIHandler) handleImageReferenceValidity(request *restful.Request, response *restful.Response) {
	spec := new(validation.ImageReferenceValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	validity, err := validation.ValidateImageReference(spec)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, validity)
}

func (apiHandler *APIHandler) handleProtocolValidity(request *restful.Request, response *restful.Response) {
	spec := new(validation.ProtocolValiditySpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, validation.ValidateProtocol(spec))
}

func (apiHandler *APIHandler) handleGetAvailableProtocols(request *restful.Request, response *restful.Response) {
	_ = response.WriteHeaderAndEntity(http.StatusOK, deployment.GetAvailableProtocols())
}

func (apiHandler *APIHandler) handleGetReplicationControllerList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicaSets(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicaSetDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	result, err := replicaset.GetReplicaSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, replicaSet)

	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicaSetPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, replicaSet, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicaSetServices(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	replicaSet := request.PathParameter("replicaSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicaset.GetReplicaSetServices(k8sClient, dataSelect, namespace, replicaSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicaSetEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicaSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)

}

func (apiHandler *APIHandler) handleGetPodEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	klog.V(4).Info("Getting events related to a pod in namespace")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := pod.GetEventsForPod(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

// Handles execute shell API call
func (apiHandler *APIHandler) handleExecShell(request *restful.Request, response *restful.Response) {
	sessionID, err := genTerminalSessionId()
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	cfg, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	terminalSessions.Set(sessionID, TerminalSession{
		id:       sessionID,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})
	go WaitForTerminal(k8sClient, cfg, request, sessionID)
	_ = response.WriteHeaderAndEntity(http.StatusOK, TerminalResponse{ID: sessionID})
}

func (apiHandler *APIHandler) handleGetDeployments(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDeploymentDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	result, err := deployment.GetDeploymentDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDeploymentEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDeploymentOldReplicaSets(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentOldReplicaSets(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDeploymentNewReplicaSet(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("deployment")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := deployment.GetDeploymentNewReplicaSet(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics // download standard metrics - cpu, and memory - by default
	result, err := pod.GetPodList(k8sClient, apiHandler.iManager.Metric().Client(), namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPodDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := pod.GetPodDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicationControllerDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	result, err := replicationcontroller.GetReplicationControllerDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleUpdateReplicasCount(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	spec := new(replicationcontroller.ReplicationControllerSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err := replicationcontroller.UpdateReplicasCount(k8sClient, namespace, name, spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func (apiHandler *APIHandler) handleGetResource(request *restful.Request, response *restful.Response) {
	verber, err := client.VerberClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	kind := request.PathParameter("kind")
	namespace := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	result, err := verber.Get(kind, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handlePutResource(
	request *restful.Request, response *restful.Response) {
	verber, err := client.VerberClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	raw := &unstructured.Unstructured{}
	bytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	err = raw.UnmarshalJSON(bytes)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	if err = verber.Update(raw); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (apiHandler *APIHandler) handleDeleteResource(
	request *restful.Request, response *restful.Response) {
	verber, err := client.VerberClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	kind := request.PathParameter("kind")
	namespace := request.PathParameters()["namespace"]
	name := request.PathParameter("name")
	propagation := request.QueryParameter("propagation")
	deleteNow := request.QueryParameter("deleteNow") == "true"

	if err := verber.Delete(kind, namespace, name, propagation, deleteNow); err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (apiHandler *APIHandler) handleGetReplicationControllerPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	rc := request.PathParameter("replicationController")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := replicationcontroller.GetReplicationControllerPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, rc, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleCreateNamespace(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespaceSpec := new(ns.NamespaceSpec)
	if err := request.ReadEntity(namespaceSpec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	if err := ns.CreateNamespace(namespaceSpec, k8sClient); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusCreated, namespaceSpec)
}

func (apiHandler *APIHandler) handleGetNamespaces(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := ns.GetNamespaceList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNamespaceDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	result, err := ns.GetNamespaceDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNamespaceEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := event.GetNamespaceEvents(k8sClient, dataSelect, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetEventList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := event.GetEventList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleCreateImagePullSecret(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	spec := new(secret.ImagePullSecretSpec)
	if err := request.ReadEntity(spec); err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	result, err := secret.CreateSecret(k8sClient, spec)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusCreated, result)
}

func (apiHandler *APIHandler) handleGetSecretDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := secret.GetSecretDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetSecretList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	namespace := parseNamespacePathParameter(request)
	result, err := secret.GetSecretList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetConfigMapList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := configmap.GetConfigMapList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetConfigMapDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("configmap")
	result, err := configmap.GetConfigMapDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPersistentVolumeList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := persistentvolume.GetPersistentVolumeList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPersistentVolumeDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("persistentvolume")
	result, err := persistentvolume.GetPersistentVolumeDetail(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPersistentVolumeClaimList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := persistentvolumeclaim.GetPersistentVolumeClaimList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPersistentVolumeClaimDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := persistentvolumeclaim.GetPersistentVolumeClaimDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPodContainers(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	result, err := container.GetPodContainers(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicationControllerEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetReplicationControllerServices(request *restful.Request,
	response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("replicationController")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := replicationcontroller.GetReplicationControllerServices(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDaemonSetList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDaemonSetDetail(
	request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	result, err := daemonset.GetDaemonSetDetail(k8sClient, apiHandler.iManager.Metric().Client(), namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDaemonSetPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := daemonset.GetDaemonSetPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, name, namespace)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDaemonSetServices(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	daemonSet := request.PathParameter("daemonSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := daemonset.GetDaemonSetServices(k8sClient, dataSelect, namespace, daemonSet)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetDaemonSetEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := event.GetResourceEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandle *APIHandler) handleDaemonSetRestart(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	result, err := daemonset.RestartDaemonSet(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandle *APIHandler) handleStatefulSetRestart(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("daemonSet")
	result, err := statefulset.RestartStatefulSet(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetHorizontalPodAutoscalerList(request *restful.Request,
	response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := horizontalpodautoscaler.GetHorizontalPodAutoscalerList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetHorizontalPodAutoscalerListForResource(request *restful.Request,
	response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	kind := request.PathParameter("kind")
	result, err := horizontalpodautoscaler.GetHorizontalPodAutoscalerListForResource(k8sClient, namespace, kind, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetHorizontalPodAutoscalerDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("horizontalpodautoscaler")
	result, err := horizontalpodautoscaler.GetHorizontalPodAutoscalerDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetJobList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobList(k8sClient, namespace, dataSelect, apiHandler.iManager.Metric().Client())
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetJobDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := job.GetJobDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetJobPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := job.GetJobPods(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetJobEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := job.GetJobEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCronJobList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.NoMetrics
	result, err := cronjob.GetCronJobList(k8sClient, namespace, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCronJobDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	result, err := cronjob.GetCronJobDetail(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCronJobJobs(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	active := true
	if request.QueryParameter("active") == "false" {
		active = false
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobJobs(k8sClient, apiHandler.iManager.Metric().Client(), dataSelect, namespace, name, active)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCronJobEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := cronjob.GetCronJobEvents(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleTriggerCronJob(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")
	err = cronjob.TriggerCronJob(k8sClient, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (apiHandler *APIHandler) handleGetStorageClassList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := storageclass.GetStorageClassList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetStorageClass(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("storageclass")
	result, err := storageclass.GetStorageClass(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetStorageClassPersistentVolumes(request *restful.Request,
	response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("storageclass")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := persistentvolume.GetStorageClassPersistentVolumes(k8sClient,
		name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetIngressClassList(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := ingressclass.GetIngressClassList(k8sClient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetIngressClass(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("ingressclass")
	result, err := ingressclass.GetIngressClass(k8sClient, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetPodPersistentVolumeClaims(request *restful.Request,
	response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("pod")
	namespace := request.PathParameter("namespace")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := persistentvolumeclaim.GetPodPersistentVolumeClaims(k8sClient,
		namespace, name, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCustomResourceDefinitionList(request *restful.Request, response *restful.Response) {
	apiextensionsclient, err := client.APIExtensionsClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceDefinitionList(apiextensionsclient, dataSelect)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCustomResourceDefinitionDetail(request *restful.Request, response *restful.Response) {
	config, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("crd")
	result, err := customresourcedefinition.GetCustomResourceDefinitionDetail(apiextensionsclient, config, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCustomResourceObjectList(request *restful.Request, response *restful.Response) {
	config, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	dataSelect := parser.ParseDataSelectPathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectList(apiextensionsclient, config, namespace, dataSelect, crdName)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCustomResourceObjectDetail(request *restful.Request, response *restful.Response) {
	config, err := client.Config(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	apiextensionsclient, err := client.APIExtensionsClient(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("object")
	crdName := request.PathParameter("crd")
	namespace := parseNamespacePathParameter(request)
	result, err := customresourcedefinition.GetCustomResourceObjectDetail(apiextensionsclient, namespace, config, crdName, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetCustomResourceObjectEvents(request *restful.Request, response *restful.Response) {
	klog.V(4).Info("Getting events related to a custom resource object in namespace")

	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("object")
	namespace := request.PathParameter("namespace")
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	result, err := customresourcedefinition.GetEventsForCustomResourceObject(k8sClient, dataSelect, namespace, name)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleLogSource(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	resourceName := request.PathParameter("resourceName")
	resourceType := request.PathParameter("resourceType")
	namespace := request.PathParameter("namespace")
	logSources, err := logs.GetLogSources(k8sClient, namespace, resourceName, resourceType)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, logSources)
}

func (apiHandler *APIHandler) handleLogs(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")

	refTimestamp := request.QueryParameter("referenceTimestamp")
	if refTimestamp == "" {
		refTimestamp = logs.NewestTimestamp
	}

	refLineNum, err := strconv.Atoi(request.QueryParameter("referenceLineNum"))
	if err != nil {
		refLineNum = 0
	}
	usePreviousLogs := request.QueryParameter("previous") == "true"
	offsetFrom, err1 := strconv.Atoi(request.QueryParameter("offsetFrom"))
	offsetTo, err2 := strconv.Atoi(request.QueryParameter("offsetTo"))
	logFilePosition := request.QueryParameter("logFilePosition")

	logSelector := logs.DefaultSelection
	if err1 == nil && err2 == nil {
		logSelector = &logs.Selection{
			ReferencePoint: logs.LogLineId{
				LogTimestamp: logs.LogTimestamp(refTimestamp),
				LineNum:      refLineNum,
			},
			OffsetFrom:      offsetFrom,
			OffsetTo:        offsetTo,
			LogFilePosition: logFilePosition,
		}
	}

	result, err := container.GetLogDetails(k8sClient, namespace, podID, containerID, logSelector, usePreviousLogs)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	_ = response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleLogFile(request *restful.Request, response *restful.Response) {
	k8sClient, err := client.Client(request.Request)
	opts := new(v1.PodLogOptions)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}

	namespace := request.PathParameter("namespace")
	podID := request.PathParameter("pod")
	containerID := request.PathParameter("container")
	opts.Previous = request.QueryParameter("previous") == "true"
	opts.Timestamps = request.QueryParameter("timestamps") == "true"

	logStream, err := container.GetLogFile(k8sClient, namespace, podID, containerID, opts)
	if err != nil {
		errors.HandleInternalError(response, err)
		return
	}
	handleDownload(response, logStream)
}

// parseNamespacePathParameter parses namespace selector for list pages in path parameter.
// The namespace selector is a comma separated list of namespaces that are trimmed.
// No namespaces mean "view all user namespaces", i.e., everything except kube-system.
func parseNamespacePathParameter(request *restful.Request) *common.NamespaceQuery {
	namespace := request.PathParameter("namespace")
	namespaces := strings.Split(namespace, ",")
	var nonEmptyNamespaces []string
	for _, n := range namespaces {
		n = strings.Trim(n, " ")
		if len(n) > 0 {
			nonEmptyNamespaces = append(nonEmptyNamespaces, n)
		}
	}
	return common.NewNamespaceQuery(nonEmptyNamespaces)
}
