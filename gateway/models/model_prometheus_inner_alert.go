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

// PrometheusInnerAlert is an object. A single alert produced by Prometheus
type PrometheusInnerAlert struct {
	// Labels: A single label of a Prometheus alert
	Labels PrometheusInnerAlertLabel `json:"labels" mapstructure:"labels"`
	// Status: The status of the alert
	Status string `json:"status" mapstructure:"status"`
}

// Validate implements basic validation for this model
func (m PrometheusInnerAlert) Validate() error {
	return validation.Errors{
		"labels": validation.Validate(
			m.Labels, validation.NotNil,
		),
	}.Filter()
}

// GetLabels returns the Labels property
func (m PrometheusInnerAlert) GetLabels() PrometheusInnerAlertLabel {
	return m.Labels
}

// SetLabels sets the Labels property
func (m *PrometheusInnerAlert) SetLabels(val PrometheusInnerAlertLabel) {
	m.Labels = val
}

// GetStatus returns the Status property
func (m PrometheusInnerAlert) GetStatus() string {
	return m.Status
}

// SetStatus sets the Status property
func (m *PrometheusInnerAlert) SetStatus(val string) {
	m.Status = val
}
