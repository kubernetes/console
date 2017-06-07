package heapster

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/kubernetes/dashboard/src/app/backend/client"
	integrationapi "github.com/kubernetes/dashboard/src/app/backend/integration/api"
	"github.com/kubernetes/dashboard/src/app/backend/integration/metric/aggregation"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	heapster "k8s.io/heapster/metrics/api/v1/types"
)

const HeapsterIntegrationID integrationapi.IntegrationID = "heapster"

// Heapster client implements metric client API.
type HeapsterClient struct {
	client HeapsterRESTClient
}

// Implement IntegrationApp interface

func (self HeapsterClient) HealthCheck() error {
	if self.client == nil {
		return errors.New("Heapster not configured")
	}

	return self.client.HealthCheck()
}

func (self HeapsterClient) ID() integrationapi.IntegrationID {
	return HeapsterIntegrationID
}

// Implement MetricClient interface

func (self HeapsterClient) DownloadMetrics(selectors []metricapi.ResourceSelector,
	metricNames []string, cachedResources *metricapi.CachedResources) metricapi.MetricPromises {
	result := metricapi.MetricPromises{}
	for _, metricName := range metricNames {
		collectedMetrics := self.DownloadMetric(selectors, metricName, cachedResources)
		result = append(result, collectedMetrics...)
	}
	return result
}

func (self HeapsterClient) DownloadMetric(selectors []metricapi.ResourceSelector,
	metricName string, cachedResources *metricapi.CachedResources) metricapi.MetricPromises {
	heapsterSelectors := getHeapsterSelectors(selectors, cachedResources)

	// Downloads metric in the fastest possible way by first compressing HeapsterSelectors and later unpacking the result to separate boxes.
	compressedSelectors, reverseMapping := compress(heapsterSelectors)
	return self.downloadMetric(heapsterSelectors, compressedSelectors, reverseMapping, metricName)
}

func (self HeapsterClient) AggregateMetrics(metrics metricapi.MetricPromises, metricName string,
	aggregations metricapi.AggregationModes) metricapi.MetricPromises {
	return aggregation.AggregateMetricPromises(metrics, metricName, aggregations, nil)
}

func (self HeapsterClient) downloadMetric(heapsterSelectors []heapsterSelector,
	compressedSelectors []heapsterSelector, reverseMapping map[string][]int,
	metricName string) metricapi.MetricPromises {
	// collect all the required data (as promises)
	unassignedResourcePromisesList := make([]metricapi.MetricPromises, len(compressedSelectors))
	for selectorId, compressedSelector := range compressedSelectors {
		unassignedResourcePromisesList[selectorId] =
			self.downloadMetricForEachTargetResource(compressedSelector, metricName)
	}
	// prepare final result
	result := metricapi.NewMetricPromises(len(heapsterSelectors))
	// unpack downloaded data - this is threading safe because there is only one thread running.
	go func() {
		// unpack the data selector by selector.
		for selectorId, selector := range compressedSelectors {
			unassignedResourcePromises := unassignedResourcePromisesList[selectorId]
			// now unpack the resources and push errors in case of error.
			unassignedResources, err := unassignedResourcePromises.GetMetrics()
			if err != nil {
				for _, originalMappingIndex := range reverseMapping[selector.Path] {
					result[originalMappingIndex].Error <- err
					result[originalMappingIndex].Metric <- nil
				}
				continue
			}
			unassignedResourceMap := map[string]metricapi.Metric{}
			for _, unassignedMetric := range unassignedResources {
				unassignedResourceMap[unassignedMetric.
					Label[selector.TargetResourceType][0]] = unassignedMetric
			}

			// now, if everything went ok, unpack the metrics into original selectors
			for _, originalMappingIndex := range reverseMapping[selector.Path] {
				// find out what resources this selector needs
				requestedResources := []metricapi.Metric{}
				for _, requestedResourceName := range heapsterSelectors[originalMappingIndex].Resources {
					requestedResources = append(requestedResources,
						unassignedResourceMap[requestedResourceName])
				}
				// aggregate the data for this resource

				aggregatedMetric := aggregation.AggregateData(requestedResources, metricName, "sum")
				aggregatedMetric.Label = heapsterSelectors[originalMappingIndex].Label
				result[originalMappingIndex].Metric <- &aggregatedMetric
				result[originalMappingIndex].Error <- nil
			}
		}
	}()
	return result
}

// downloadMetricForEachTargetResource downloads requested metric for each resource present in HeapsterSelector
// and returns the result as a list of promises - one promise for each resource. Order of promises returned is the same as order in self.Resources.
func (self HeapsterClient) downloadMetricForEachTargetResource(selector heapsterSelector, metricName string) metricapi.MetricPromises {
	var notAggregatedMetrics metricapi.MetricPromises
	if HeapsterAllInOneDownloadConfig[selector.TargetResourceType] {
		notAggregatedMetrics = self.allInOneDownload(selector, metricName)
	} else {
		notAggregatedMetrics = metricapi.MetricPromises{}
		for i := range selector.Resources {
			notAggregatedMetrics = append(notAggregatedMetrics, self.ithResourceDownload(selector, metricName, i))
		}
	}
	return notAggregatedMetrics
}

// ithResourceDownload downloads metric for ith resource in self.Resources. Use only in case all in 1 download is not supported
// for this resource type.
func (self HeapsterClient) ithResourceDownload(selector heapsterSelector, metricName string,
	i int) metricapi.MetricPromise {
	result := metricapi.NewMetricPromise()
	go func() {
		rawResult := heapster.MetricResult{}
		err := self.unmarshalType(selector.Path+selector.Resources[i]+"/metrics/"+metricName, &rawResult)
		if err != nil {
			result.Metric <- nil
			result.Error <- err
			return
		}
		dataPoints := DataPointsFromMetricJSONFormat(rawResult)

		result.Metric <- &metricapi.Metric{
			DataPoints:   dataPoints,
			MetricPoints: toMetricPoints(rawResult.Metrics),
			MetricName:   metricName,
			Label: metricapi.Label{
				selector.TargetResourceType: []string{selector.Resources[i]},
			},
		}
		result.Error <- nil
		return
	}()
	return result
}

// allInOneDownload downloads metrics for all resources present in self.Resources in one request.
// returns a list of metric promises - one promise for each resource. Order of self.Resources is preserved.
func (self HeapsterClient) allInOneDownload(selector heapsterSelector, metricName string) metricapi.MetricPromises {
	result := metricapi.NewMetricPromises(len(selector.Resources))
	go func() {
		if len(selector.Resources) == 0 {
			return
		}
		rawResults := heapster.MetricResultList{}
		err := self.unmarshalType(selector.Path+strings.Join(selector.Resources, ",")+"/metrics/"+metricName, &rawResults)
		if err != nil {
			result.PutMetrics(nil, err)
			return
		}
		if len(result) != len(rawResults.Items) {
			result.PutMetrics(nil, fmt.Errorf(`Received invalid number of resources from heapster. Expected %d received %d`, len(result), len(rawResults.Items)))
			return
		}

		for i, rawResult := range rawResults.Items {
			dataPoints := DataPointsFromMetricJSONFormat(rawResult)

			result[i].Metric <- &metricapi.Metric{
				DataPoints:   dataPoints,
				MetricPoints: toMetricPoints(rawResult.Metrics),
				MetricName:   metricName,
				Label: metricapi.Label{
					selector.TargetResourceType: []string{selector.Resources[i]},
				},
			}
			result[i].Error <- nil
		}
		return

	}()
	return result
}

// unmarshalType performs heapster GET request to the specifies path and transfers
// the data to the interface provided.
func (self HeapsterClient) unmarshalType(path string, v interface{}) error {
	rawData, err := self.client.Get("/model/" + path).DoRaw()
	if err != nil {
		return err
	}
	return json.Unmarshal(rawData, v)
}

// CreateHeapsterClient creates new Heapster client. When heapsterHost param is empty
// string the function assumes that it is running inside a Kubernetes cluster and connects via
// service proxy. heapsterHost param is in the format of protocol://address:port,
// e.g., http://localhost:8002.
func CreateHeapsterClient(host string, k8sClient *kubernetes.Clientset) (
	metricapi.MetricClient, error) {

	if host == "" {
		log.Print("Creating in-cluster Heapster client")
		c := inClusterHeapsterClient{client: k8sClient.Core().RESTClient()}
		return HeapsterClient{client: c}, nil
	}

	cfg := &rest.Config{Host: host, QPS: client.DefaultQPS, Burst: client.DefaultBurst}
	restClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return HeapsterClient{}, err
	}
	log.Printf("Creating remote Heapster client for %s", host)
	c := remoteHeapsterClient{client: restClient.Core().RESTClient()}
	return HeapsterClient{client: c}, nil
}
