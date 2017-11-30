// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
