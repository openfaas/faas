package statsd

import (
	"fmt"
	"time"

	dogstatsd "github.com/DataDog/datadog-go/statsd"
	"github.com/openfaas/faas/gateway/metrics"
)

type Client struct {
	client *dogstatsd.Client
}

func NewClient(statsDServer string) (metrics.Metrics, error) {
	client, err := dogstatsd.New(statsDServer)
	if err != nil {
		return nil, fmt.Errorf("Unable to create statsd client %s", err.Error())
	}

	client.Namespace = "openfaas.gateway."

	return &Client{client}, nil
}

func (c *Client) GatewayFunctionInvocation(labels map[string]string) {
	c.client.Incr(
		"function_invocation",
		convertMapToDataDogLabels(labels),
		1,
	)
}

func (c *Client) GatewayFunctionsHistogram(labels map[string]string, duration time.Duration) {
	c.client.Timing(
		"function_invocation",
		duration,
		convertMapToDataDogLabels(labels),
		1,
	)
}

func (c *Client) ServiceReplicasCounter(labels map[string]string, replicas float64) {
	c.client.Gauge(
		"replicas",
		replicas,
		convertMapToDataDogLabels(labels),
		1,
	)
}

func convertMapToDataDogLabels(m map[string]string) []string {
	labels := make([]string, 0)

	for k, v := range m {
		labels = append(labels, fmt.Sprintf("%s:%s", k, v))
	}

	return labels
}
