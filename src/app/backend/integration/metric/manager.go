package metric

import (
	"fmt"

	"github.com/emicklei/go-restful/log"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	integrationapi "github.com/kubernetes/dashboard/src/app/backend/integration/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/integration/metric/heapster"
	"k8s.io/client-go/kubernetes"
)

type MetricManager interface {
	// Returns active Metric client.
	Client() metricapi.MetricClient
	Enable(integrationapi.IntegrationID) error
	List() []integrationapi.Integration

	ConfigureHeapster(host string, client *kubernetes.Clientset) MetricManager
}

type metricManager struct {
	manager client.ClientManager
	clients map[integrationapi.IntegrationID]metricapi.MetricClient
	active  metricapi.MetricClient
}

func (self *metricManager) Client() metricapi.MetricClient {
	return self.active
}

func (self *metricManager) Enable(id integrationapi.IntegrationID) error {
	metricClient, exists := self.clients[id]
	if !exists {
		return fmt.Errorf("No metric client found for integration id: %s", id)
	}

	err := metricClient.HealthCheck()
	if err != nil {
		return fmt.Errorf("Health check failed: %s", err.Error())
	}

	self.active = metricClient
	return nil
}

func (self *metricManager) List() []integrationapi.Integration {
	result := make([]integrationapi.Integration, 0)
	for _, c := range self.clients {
		result = append(result, c.(integrationapi.Integration))
	}

	return result
}

func (self *metricManager) ConfigureHeapster(host string,
	client *kubernetes.Clientset) MetricManager {
	kubeClient, err := self.manager.Client(nil)
	if err != nil {
		log.Print(err)
		return self
	}

	metricClient, err := heapster.CreateHeapsterClient(host, kubeClient)
	if err != nil {
		log.Printf("There was an error during heapster client creation: %s", err.Error())
		return self
	}

	self.clients[metricClient.ID()] = metricClient
	return self
}

func NewMetricManager(manager client.ClientManager) MetricManager {
	return &metricManager{
		manager: manager,
		clients: make(map[integrationapi.IntegrationID]metricapi.MetricClient),
	}
}
