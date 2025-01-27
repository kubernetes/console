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

package deployment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	apps "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/api/pkg/args"
	derrors "k8s.io/dashboard/errors"
)

const (
	// DescriptionAnnotationKey is annotation key for a description.
	DescriptionAnnotationKey = "description"
)

// AppDeploymentSpec is a specification for an app deployment.
type AppDeploymentSpec struct {
	// Name of the application.
	Name string `json:"name"`

	// Docker image path for the application.
	ContainerImage string `json:"containerImage"`

	// The name of an image pull secret in case of a private docker repository.
	ImagePullSecret *string `json:"imagePullSecret"`

	// Command that is executed instead of container entrypoint, if specified.
	ContainerCommand *string `json:"containerCommand"`

	// Arguments for the specified container command or container entrypoint (if command is not
	// specified here).
	ContainerCommandArgs *string `json:"containerCommandArgs"`

	// Number of replicas of the image to maintain.
	Replicas int32 `json:"replicas"`

	// Port mappings for the service that is created. The service is created if there is at least
	// one port mapping.
	PortMappings []PortMapping `json:"portMappings"`

	// List of user-defined environment variables.
	Variables []EnvironmentVariable `json:"variables"`

	// Whether the created service is external.
	IsExternal bool `json:"isExternal"`

	// Description of the deployment.
	Description *string `json:"description"`

	// Target namespace of the application.
	Namespace string `json:"namespace"`

	// Optional memory requirement for the container.
	MemoryRequirement *resource.Quantity `json:"memoryRequirement"`

	// Optional CPU requirement for the container.
	CpuRequirement *resource.Quantity `json:"cpuRequirement"`

	// Labels that will be defined on Pods/RCs/Services
	Labels []Label `json:"labels"`

	// Whether to run the container as privileged user (essentially equivalent to root on the host).
	RunAsPrivileged bool `json:"runAsPrivileged"`
}

// AppDeploymentFromFileSpec is a specification for deployment from file
type AppDeploymentFromFileSpec struct {
	// Name of the file
	Name string `json:"name"`

	// Namespace that object should be deployed in
	Namespace string `json:"namespace"`

	// File content
	Content string `json:"content"`

	// Whether validate content before creation or not
	Validate bool `json:"validate"`
}

// AppDeploymentFromFileResponse is a specification for deployment from file
type AppDeploymentFromFileResponse struct {
	// Name of the file
	Name string `json:"name"`

	// File content
	Content string `json:"content"`

	// Error after create resource
	Error string `json:"error"`
}

// PortMapping is a specification of port mapping for an application deployment.
type PortMapping struct {
	// Port that will be exposed on the service.
	Port int32 `json:"port"`

	// Docker image path for the application.
	TargetPort int32 `json:"targetPort"`

	// IP protocol for the mapping, e.g., "TCP" or "UDP".
	Protocol api.Protocol `json:"protocol"`
}

// EnvironmentVariable represents a named variable accessible for containers.
type EnvironmentVariable struct {
	// Name of the variable. Must be a C_IDENTIFIER.
	Name string `json:"name"`

	// Value of the variable, as defined in Kubernetes core API.
	Value string `json:"value"`
}

// Label is a structure representing label assignable to Pod/RC/Service
type Label struct {
	// Label key
	Key string `json:"key"`

	// Label value
	Value string `json:"value"`
}

// Protocols is a structure representing supported protocol types for a service
type Protocols struct {
	// Array containing supported protocol types e.g., ["TCP", "UDP"]
	Protocols []api.Protocol `json:"protocols"`
}

// DeployApp deploys an app based on the given configuration. The app is deployed using the given
// client. App deployment consists of a deployment and an optional service. Both of them
// share common labels.
func DeployApp(spec *AppDeploymentSpec, client client.Interface) error {
	klog.V(args.LogLevelVerbose).Infof("Deploying %s application into %s namespace", spec.Name, spec.Namespace)

	annotations := map[string]string{}
	if spec.Description != nil {
		annotations[DescriptionAnnotationKey] = *spec.Description
	}
	labels := getLabelsMap(spec.Labels)
	objectMeta := metaV1.ObjectMeta{
		Annotations: annotations,
		Name:        spec.Name,
		Labels:      labels,
	}

	containerSpec := api.Container{
		Name:  spec.Name,
		Image: spec.ContainerImage,
		SecurityContext: &api.SecurityContext{
			Privileged: &spec.RunAsPrivileged,
		},
		Resources: api.ResourceRequirements{
			Requests: make(map[api.ResourceName]resource.Quantity),
		},
		Env: convertEnvVarsSpec(spec.Variables),
	}

	if spec.ContainerCommand != nil {
		containerSpec.Command = []string{*spec.ContainerCommand}
	}
	if spec.ContainerCommandArgs != nil {
		containerSpec.Args = strings.Fields(*spec.ContainerCommandArgs)
	}

	if spec.CpuRequirement != nil {
		containerSpec.Resources.Requests[api.ResourceCPU] = *spec.CpuRequirement
	}
	if spec.MemoryRequirement != nil {
		containerSpec.Resources.Requests[api.ResourceMemory] = *spec.MemoryRequirement
	}
	podSpec := api.PodSpec{
		Containers: []api.Container{containerSpec},
	}
	if spec.ImagePullSecret != nil {
		podSpec.ImagePullSecrets = []api.LocalObjectReference{{Name: *spec.ImagePullSecret}}
	}

	podTemplate := api.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec:       podSpec,
	}

	deployment := &apps.Deployment{
		ObjectMeta: objectMeta,
		Spec: apps.DeploymentSpec{
			Replicas: &spec.Replicas,
			Template: podTemplate,
			Selector: &metaV1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
	_, err := client.AppsV1().Deployments(spec.Namespace).Create(context.TODO(), deployment, metaV1.CreateOptions{})

	if err != nil {
		return err
	}

	if len(spec.PortMappings) > 0 {
		service := &api.Service{
			ObjectMeta: objectMeta,
			Spec: api.ServiceSpec{
				Selector: labels,
			},
		}

		if spec.IsExternal {
			service.Spec.Type = api.ServiceTypeLoadBalancer
		} else {
			service.Spec.Type = api.ServiceTypeClusterIP
		}

		for _, portMapping := range spec.PortMappings {
			servicePort :=
				api.ServicePort{
					Protocol: portMapping.Protocol,
					Port:     portMapping.Port,
					Name:     generatePortMappingName(portMapping),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: portMapping.TargetPort,
					},
				}
			service.Spec.Ports = append(service.Spec.Ports, servicePort)
		}

		_, err = client.CoreV1().Services(spec.Namespace).Create(context.TODO(), service, metaV1.CreateOptions{})
		return err
	}

	return nil
}

// GetAvailableProtocols returns list of available protocols. Currently it is TCP and UDP.
func GetAvailableProtocols() *Protocols {
	return &Protocols{Protocols: []api.Protocol{api.ProtocolTCP, api.ProtocolUDP}}
}

func convertEnvVarsSpec(variables []EnvironmentVariable) []api.EnvVar {
	var result []api.EnvVar
	for _, variable := range variables {
		result = append(result, api.EnvVar{Name: variable.Name, Value: variable.Value})
	}
	return result
}

func generatePortMappingName(portMapping PortMapping) string {
	return generateName(fmt.Sprintf("%s-%d-%d-", strings.ToLower(string(portMapping.Protocol)),
		portMapping.Port, portMapping.TargetPort))
}

func generateName(base string) string {
	maxNameLength := 63
	randomLength := 5
	maxGeneratedNameLength := maxNameLength - randomLength
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, rand.String(randomLength))
}

// Converts array of labels to map[string]string
func getLabelsMap(labels []Label) map[string]string {
	result := make(map[string]string)

	for _, label := range labels {
		result[label.Key] = label.Value
	}

	return result
}

// DeployAppFromFile deploys an app based on the given yaml or json file.
func DeployAppFromFile(cfg *rest.Config, spec *AppDeploymentFromFileSpec) (bool, error) {
	reader := strings.NewReader(spec.Content)
	klog.V(args.LogLevelVerbose).Infof("Namespace for deploy from file: %s\n", spec.Namespace)
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		data := &unstructured.Unstructured{}
		if err := d.Decode(data); err != nil {
			if errors.Is(err, io.EOF) {
				return true, nil
			}
			return false, err
		}

		version := data.GetAPIVersion()
		kind := data.GetKind()

		gv, err := schema.ParseGroupVersion(version)
		if err != nil {
			gv = schema.GroupVersion{Version: version}
		}

		discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
		if err != nil {
			return false, err
		}

		apiResourceList, err := discoveryClient.ServerResourcesForGroupVersion(version)
		if err != nil {
			return false, err
		}
		apiResources := apiResourceList.APIResources
		var resource *metaV1.APIResource
		for _, apiResource := range apiResources {
			if apiResource.Kind == kind && !strings.Contains(apiResource.Name, "/") {
				resource = &apiResource
				break
			}
		}
		if resource == nil {
			return false, fmt.Errorf("unknown resource kind: %s", kind)
		}

		dynamicClient, err := dynamic.NewForConfig(cfg)
		if err != nil {
			return false, err
		}

		groupVersionResource := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
		namespace := spec.Namespace

		if strings.Compare(spec.Namespace, "_all") == 0 {
			namespace = data.GetNamespace()
		}

		if namespace == "" {
			namespace = "default"
		}

		if resource.Namespaced {
			_, err = dynamicClient.Resource(groupVersionResource).Namespace(namespace).Create(context.TODO(), data, metaV1.CreateOptions{})
		} else {
			_, err = dynamicClient.Resource(groupVersionResource).Create(context.TODO(), data, metaV1.CreateOptions{})
		}

		if err != nil {
			return false, derrors.LocalizeError(err)
		}
	}
}
