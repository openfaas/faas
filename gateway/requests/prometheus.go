// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) Alex Ellis 2017. All rights reserved.

package requests

// PrometheusInnerAlertLabel PrometheusInnerAlertLabel
type PrometheusInnerAlertLabel struct {
	AlertName    string `json:"alertname"`
	FunctionName string `json:"function_name"`
}

// PrometheusInnerAlert PrometheusInnerAlert
type PrometheusInnerAlert struct {
	Status string                    `json:"status"`
	Labels PrometheusInnerAlertLabel `json:"labels"`
}

// PrometheusAlert as produced by AlertManager
type PrometheusAlert struct {
	Status   string                 `json:"status"`
	Receiver string                 `json:"receiver"`
	Alerts   []PrometheusInnerAlert `json:"alerts"`
}
