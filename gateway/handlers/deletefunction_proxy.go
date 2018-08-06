package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/openfaas/faas/gateway/types"
)

// MakeDeleteFunctionProxyHandler creates a handler which forwards HTTP requests and sets replicasCount to 0
func MakeDeleteFunctionProxyHandler(proxy *types.HTTPClientReverseProxy, notifiers []HTTPNotifier, baseURLResolver BaseURLResolver, metricsOptions metrics.MetricOptions) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		functionName, parseErr := getFunctionNameFromRequest(r)
		if parseErr != nil {
			log.Printf("Error parsing request to remove service.")
		}

		baseURL := baseURLResolver.Resolve(r)

		requestURL := r.URL.Path

		start := time.Now()

		statusCode, err := forwardRequest(w, r, proxy.Client, baseURL, requestURL, proxy.Timeout)

		seconds := time.Since(start)

		if err != nil {
			log.Printf("Error with upstream request to: %s, %s\n", requestURL, err.Error())
		} else {
			log.Printf("Setting replicas count to 0 for %s\n", functionName)
			trackFunctionStop(metricsOptions, functionName)
		}
		for _, notifier := range notifiers {
			notifier.Notify(r.Method, requestURL, statusCode, seconds)
		}
	}
}

func getFunctionNameFromRequest(r *http.Request) (string, error) {
	req := requests.DeleteFunctionRequest{}
	reqData, _ := ioutil.ReadAll(r.Body)

	unmarshalErr := json.Unmarshal(reqData, &req)
	if unmarshalErr != nil {
		log.Printf("Error unmarshaling the request: %s\n", unmarshalErr.Error())
		return "", unmarshalErr
	} else if len(req.FunctionName) == 0 {
		log.Printf("The function name in the request was empty\n")
		return "", fmt.Errorf("Function name empty")
	}

	// Restore the io.ReadCloser to its original state so the body is readable
	// when forwarding the request
	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqData))

	return req.FunctionName, nil
}
