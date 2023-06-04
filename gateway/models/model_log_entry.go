// This file is auto-generated, DO NOT EDIT.
//
// Source:
//
//	Title: OpenFaaS API Gateway
//	Version: 0.8.12
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"time"
)

// LogEntry is an object.
type LogEntry struct {
	// Instance: the name/id of the specific function instance
	Instance string `json:"instance" mapstructure:"instance"`
	// Name: the function name
	Name string `json:"name" mapstructure:"name"`
	// Namespace: the namespace of the function
	Namespace string `json:"namespace" mapstructure:"namespace"`
	// Text: raw log message content
	Text string `json:"text" mapstructure:"text"`
	// Timestamp: the timestamp of when the log message was recorded
	Timestamp time.Time `json:"timestamp" mapstructure:"timestamp"`
}

// Validate implements basic validation for this model
func (m LogEntry) Validate() error {
	return validation.Errors{}.Filter()
}

// GetInstance returns the Instance property
func (m LogEntry) GetInstance() string {
	return m.Instance
}

// SetInstance sets the Instance property
func (m *LogEntry) SetInstance(val string) {
	m.Instance = val
}

// GetName returns the Name property
func (m LogEntry) GetName() string {
	return m.Name
}

// SetName sets the Name property
func (m *LogEntry) SetName(val string) {
	m.Name = val
}

// GetNamespace returns the Namespace property
func (m LogEntry) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *LogEntry) SetNamespace(val string) {
	m.Namespace = val
}

// GetText returns the Text property
func (m LogEntry) GetText() string {
	return m.Text
}

// SetText sets the Text property
func (m *LogEntry) SetText(val string) {
	m.Text = val
}

// GetTimestamp returns the Timestamp property
func (m LogEntry) GetTimestamp() time.Time {
	return m.Timestamp
}

// SetTimestamp sets the Timestamp property
func (m *LogEntry) SetTimestamp(val time.Time) {
	m.Timestamp = val
}
