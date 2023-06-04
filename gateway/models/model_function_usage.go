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

// FunctionUsage is an object.
type FunctionUsage struct {
	// Cpu: is the increase in CPU usage since the last measurement
	// equivalent to Kubernetes' concept of millicores.
	Cpu float64 `json:"cpu,omitempty" mapstructure:"cpu,omitempty"`
	// TotalMemoryBytes: is the total memory usage in bytes.
	TotalMemoryBytes float64 `json:"totalMemoryBytes,omitempty" mapstructure:"totalMemoryBytes,omitempty"`
}

// Validate implements basic validation for this model
func (m FunctionUsage) Validate() error {
	return validation.Errors{}.Filter()
}

// GetCpu returns the Cpu property
func (m FunctionUsage) GetCpu() float64 {
	return m.Cpu
}

// SetCpu sets the Cpu property
func (m *FunctionUsage) SetCpu(val float64) {
	m.Cpu = val
}

// GetTotalMemoryBytes returns the TotalMemoryBytes property
func (m FunctionUsage) GetTotalMemoryBytes() float64 {
	return m.TotalMemoryBytes
}

// SetTotalMemoryBytes sets the TotalMemoryBytes property
func (m *FunctionUsage) SetTotalMemoryBytes(val float64) {
	m.TotalMemoryBytes = val
}
