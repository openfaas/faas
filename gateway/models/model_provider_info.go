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

// ProviderInfo is an object.
type ProviderInfo struct {
	// Orchestration:
	Orchestration string `json:"orchestration" mapstructure:"orchestration"`
	// Provider: The orchestration provider / implementation
	Provider string `json:"provider" mapstructure:"provider"`
	// Version: The version of the provider
	Version *VersionInfo `json:"version" mapstructure:"version"`
}

// Validate implements basic validation for this model
func (m ProviderInfo) Validate() error {
	return validation.Errors{
		"version": validation.Validate(
			m.Version,
		),
	}.Filter()
}

// GetOrchestration returns the Orchestration property
func (m ProviderInfo) GetOrchestration() string {
	return m.Orchestration
}

// SetOrchestration sets the Orchestration property
func (m *ProviderInfo) SetOrchestration(val string) {
	m.Orchestration = val
}

// GetProvider returns the Provider property
func (m ProviderInfo) GetProvider() string {
	return m.Provider
}

// SetProvider sets the Provider property
func (m *ProviderInfo) SetProvider(val string) {
	m.Provider = val
}

// GetVersion returns the Version property
func (m ProviderInfo) GetVersion() *VersionInfo {
	return m.Version
}

// SetVersion sets the Version property
func (m *ProviderInfo) SetVersion(val *VersionInfo) {
	m.Version = val
}
