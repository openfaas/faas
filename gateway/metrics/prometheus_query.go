package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PrometheusQuery represents parameters for querying Prometheus
type PrometheusQuery struct {
	host             string
	port             int
	client           *http.Client
	userAgentVersion string
}

type PrometheusQueryFetcher interface {
	Fetch(query string) (*VectorQueryResponse, error)
}

// NewPrometheusQuery create a NewPrometheusQuery
func NewPrometheusQuery(host string, port int, client *http.Client, userAgentVersion string) PrometheusQuery {
	return PrometheusQuery{
		client:           client,
		host:             host,
		port:             port,
		userAgentVersion: userAgentVersion,
	}
}

// Fetch queries aggregated stats
func (q PrometheusQuery) Fetch(query string) (*VectorQueryResponse, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/api/v1/query?query=%s", q.host, q.port, query), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("openfaas-gateway/%s (Prometheus query)", q.userAgentVersion))

	res, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from Prometheus want: %d, got: %d, body: %s", http.StatusOK, res.StatusCode, string(bytesOut))
	}

	var values VectorQueryResponse

	if err := json.Unmarshal(bytesOut, &values); err != nil {
		return nil, fmt.Errorf("error unmarshaling result: %s, '%s'", err, string(bytesOut))
	}

	return &values, nil
}

type VectorQueryResponse struct {
	Data struct {
		Result []struct {
			Metric struct {
				Code         string `json:"code"`
				FunctionName string `json:"function_name"`
			}
			Value []interface{} `json:"value"`
		}
	}
}
