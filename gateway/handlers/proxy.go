// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"strconv"
	"time"

	"github.com/openfaas/faas/gateway/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func trackInvocation(service string, metrics metrics.MetricOptions, code int) {
	metrics.GatewayFunctionInvocation.With(
		prometheus.Labels{"function_name": service,
			"code": strconv.Itoa(code)}).Inc()
}

func trackTime(then time.Time, metrics metrics.MetricOptions, name string) {
	since := time.Since(then)
	metrics.GatewayFunctionsHistogram.
		WithLabelValues(name).
		Observe(since.Seconds())
}

func trackTimeExact(duration time.Duration, metrics metrics.MetricOptions, name string) {
	metrics.GatewayFunctionsHistogram.
		WithLabelValues(name).
		Observe(float64(duration))
}
