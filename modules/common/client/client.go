package client

import (
	"context"
	"net/http"
	"os"

	v1 "k8s.io/api/authorization/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func InClusterClient() client.Interface {
	if inClusterClient != nil {
		return inClusterClient
	}

	// init on-demand only
	c, err := client.NewForConfig(baseConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}

	// initialize in-memory client
	inClusterClient = c
	return inClusterClient
}

func Client(request *http.Request) (client.Interface, error) {
	return clientFromRequest(request)
}

func APIExtensionsClient(request *http.Request) (*apiextensionsclientset.Clientset, error) {
	config, err := configFromRequest(request)
	if err != nil {
		return nil, err
	}

	return apiextensionsclientset.NewForConfig(config)
}

func Config(request *http.Request) (*rest.Config, error) {
	return configFromRequest(request)
}

func CanI(request *http.Request, ssar *v1.SelfSubjectAccessReview) bool {
	k8sClient, err := Client(request)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes client")
		return false
	}

	response, err := k8sClient.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), ssar, metaV1.CreateOptions{})
	if err != nil {
		klog.ErrorS(err, "Could not create SelfSubjectAccessReview")
		return false
	}

	return response.Status.Allowed
}

func RESTClient(config *rest.Config, group, version string) (*rest.RESTClient, error) {
	groupVersion := schema.GroupVersion{
		Group:   group,
		Version: version,
	}

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupVersion,
				&metaV1.ListOptions{},
				&metaV1.DeleteOptions{},
			)
			return nil
		})
	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}

	config.GroupVersion = &groupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	return rest.RESTClientFor(config)
}
