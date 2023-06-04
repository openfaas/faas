// This file is auto-generated, DO NOT EDIT.
//
// Source:
//
//	Title: OpenFaaS API Gateway
//	Version: 0.8.12
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// PrometheusInnerAlertLabel is an object. A single label of a Prometheus alert
type PrometheusInnerAlertLabel struct {
	// Alertname: The name of the alert
	Alertname string `json:"alertname" mapstructure:"alertname"`
	// FunctionName: The name of the function
	FunctionName string `json:"function_name" mapstructure:"function_name"`
}

// Validate implements basic validation for this model
func (m PrometheusInnerAlertLabel) Validate() error {
	return validation.Errors{}.Filter()
}

// GetAlertname returns the Alertname property
func (m PrometheusInnerAlertLabel) GetAlertname() string {
	return m.Alertname
}

// SetAlertname sets the Alertname property
func (m *PrometheusInnerAlertLabel) SetAlertname(val string) {
	m.Alertname = val
}

// GetFunctionName returns the FunctionName property
func (m PrometheusInnerAlertLabel) GetFunctionName() string {
	return m.FunctionName
}

// SetFunctionName sets the FunctionName property
func (m *PrometheusInnerAlertLabel) SetFunctionName(val string) {
	m.FunctionName = val
}
