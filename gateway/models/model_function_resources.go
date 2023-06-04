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

// FunctionResources is an object.
type FunctionResources struct {
	// Cpu: The amount of cpu that is allocated for the function
	Cpu string `json:"cpu,omitempty" mapstructure:"cpu,omitempty"`
	// Memory: The amount of memory that is allocated for the function
	Memory string `json:"memory,omitempty" mapstructure:"memory,omitempty"`
}

// Validate implements basic validation for this model
func (m FunctionResources) Validate() error {
	return validation.Errors{}.Filter()
}

// GetCpu returns the Cpu property
func (m FunctionResources) GetCpu() string {
	return m.Cpu
}

// SetCpu sets the Cpu property
func (m *FunctionResources) SetCpu(val string) {
	m.Cpu = val
}

// GetMemory returns the Memory property
func (m FunctionResources) GetMemory() string {
	return m.Memory
}

// SetMemory sets the Memory property
func (m *FunctionResources) SetMemory(val string) {
	m.Memory = val
}
