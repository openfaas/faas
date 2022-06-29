package handlers

import (
	"fmt"
	"net/http"

	"golang.org/x/sync/singleflight"

	"github.com/openfaas/faas/gateway/pkg/middleware"
	"github.com/openfaas/faas/gateway/probing"
)

func MakeProbeHandler(prober probing.FunctionProber, cache probing.ProbeCacher, resolver middleware.BaseURLResolver, next http.HandlerFunc, defaultNamespace string) http.HandlerFunc {

	group := singleflight.Group{}

	return func(w http.ResponseWriter, r *http.Request) {
		functionName, namespace := middleware.GetNamespace(defaultNamespace, middleware.GetServiceName(r.URL.String()))

		key := fmt.Sprintf("Probe-%s.%s", functionName, namespace)
		res, _, _ := group.Do(key, func() (interface{}, error) {

			cached, hit := cache.Get(functionName, namespace)
			var probeResult probing.FunctionProbeResult
			if hit && cached != nil && cached.Available {
				probeResult = *cached
			} else {
				probeResult = prober.Probe(functionName, namespace)
				cache.Set(functionName, namespace, &probeResult)
			}

			return probeResult, nil
		})

		fnRes := res.(probing.FunctionProbeResult)

		if !fnRes.Available {
			http.Error(w, fmt.Sprintf("unable to probe function endpoint %s", fnRes.Error),
				http.StatusServiceUnavailable)
			return
		}

		next(w, r)
	}
}
