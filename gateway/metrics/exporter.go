// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"log"

	"github.com/openfaas/faas/gateway/requests"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is a prometheus exporter
type Exporter struct {
	metricOptions MetricOptions
	services      []requests.Function
}

// NewExporter creates a new exporter for the OpenFaaS gateway metrics
func NewExporter(options MetricOptions) *Exporter {
	return &Exporter{
		metricOptions: options,
		services:      []requests.Function{},
	}
}

// Describe is to describe the metrics for Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.metricOptions.GatewayFunctionInvocation.Describe(ch)
	e.metricOptions.GatewayFunctionsHistogram.Describe(ch)
	e.metricOptions.ServiceReplicasCounter.Describe(ch)
}

// Collect collects data to be consumed by prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.metricOptions.GatewayFunctionInvocation.Collect(ch)
	e.metricOptions.GatewayFunctionsHistogram.Collect(ch)

	e.metricOptions.ServiceReplicasCounter.Reset()
	for _, service := range e.services {
		e.metricOptions.ServiceReplicasCounter.
			WithLabelValues(service.Name).
			Set(float64(service.Replicas))
	}
	e.metricOptions.ServiceReplicasCounter.Collect(ch)
}

// StartServiceWatcher starts a ticker and collects service replica counts to expose to prometheus
func (e *Exporter) StartServiceWatcher(endpointURL url.URL, metricsOptions MetricOptions, label string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	go func() {
		for {
			select {
			case <-ticker.C:

				get, _ := http.NewRequest(http.MethodGet, endpointURL.String()+"system/functions", nil)

				services := []requests.Function{}
				res, err := proxyClient.Do(get)
				if err != nil {
					log.Println(err)
					continue
				}
				bytesOut, readErr := ioutil.ReadAll(res.Body)
				if readErr != nil {
					log.Println(err)
					continue
				}
				unmarshalErr := json.Unmarshal(bytesOut, &services)
				if unmarshalErr != nil {
					log.Println(err)
					continue
				}

				e.services = services

				break
			case <-quit:
				return
			}
		}
	}()
}