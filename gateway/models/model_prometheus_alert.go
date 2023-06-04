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

// PrometheusAlert is an object. Prometheus alert produced by AlertManager. This is only a subset of the full alert payload.
type PrometheusAlert struct {
	// Alerts: The list of alerts
	Alerts []PrometheusInnerAlert `json:"alerts" mapstructure:"alerts"`
	// Receiver: The name of the receiver
	Receiver string `json:"receiver" mapstructure:"receiver"`
	// Status: The status of the alert
	Status string `json:"status" mapstructure:"status"`
}

// Validate implements basic validation for this model
func (m PrometheusAlert) Validate() error {
	return validation.Errors{
		"alerts": validation.Validate(
			m.Alerts, validation.NotNil,
		),
	}.Filter()
}

// GetAlerts returns the Alerts property
func (m PrometheusAlert) GetAlerts() []PrometheusInnerAlert {
	return m.Alerts
}

// SetAlerts sets the Alerts property
func (m *PrometheusAlert) SetAlerts(val []PrometheusInnerAlert) {
	m.Alerts = val
}

// GetReceiver returns the Receiver property
func (m PrometheusAlert) GetReceiver() string {
	return m.Receiver
}

// SetReceiver sets the Receiver property
func (m *PrometheusAlert) SetReceiver(val string) {
	m.Receiver = val
}

// GetStatus returns the Status property
func (m PrometheusAlert) GetStatus() string {
	return m.Status
}

// SetStatus sets the Status property
func (m *PrometheusAlert) SetStatus(val string) {
	m.Status = val
}
