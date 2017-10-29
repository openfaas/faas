package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// PrometheusQuery a PrometheusQuery
type PrometheusQuery struct {
	Port   int
	Host   string
	Client *http.Client
}

type PrometheusQueryFetcher interface {
	Fetch(query string) (*VectorQueryResponse, error)
}

// NewPrometheusQuery create a NewPrometheusQuery
func NewPrometheusQuery(host string, port int, client *http.Client) PrometheusQuery {
	return PrometheusQuery{
		Client: client,
		Host:   host,
		Port:   port,
	}
}

// Fetch queries aggregated stats
func (q PrometheusQuery) Fetch(query string) (*VectorQueryResponse, error) {

	req, reqErr := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/query/?query=%s", q.Host, q.Port, query), nil)
	if reqErr != nil {
		return nil, reqErr
	}
	res, getErr := q.Client.Do(req)
	if getErr != nil {
		return nil, getErr
	}
	defer res.Body.Close()
	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	var values VectorQueryResponse

	unmarshalErr := json.Unmarshal(bytesOut, &values)
	if unmarshalErr != nil {
		return nil, unmarshalErr
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
