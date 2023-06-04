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

// GatewayInfo is an object.
type GatewayInfo struct {
	// Arch: Platform architecture
	Arch string `json:"arch" mapstructure:"arch"`
	// Provider:
	Provider *ProviderInfo `json:"provider" mapstructure:"provider"`
	// Version: version of the gateway
	Version *VersionInfo `json:"version" mapstructure:"version"`
}

// Validate implements basic validation for this model
func (m GatewayInfo) Validate() error {
	return validation.Errors{
		"provider": validation.Validate(
			m.Provider,
		),
		"version": validation.Validate(
			m.Version,
		),
	}.Filter()
}

// GetArch returns the Arch property
func (m GatewayInfo) GetArch() string {
	return m.Arch
}

// SetArch sets the Arch property
func (m *GatewayInfo) SetArch(val string) {
	m.Arch = val
}

// GetProvider returns the Provider property
func (m GatewayInfo) GetProvider() *ProviderInfo {
	return m.Provider
}

// SetProvider sets the Provider property
func (m *GatewayInfo) SetProvider(val *ProviderInfo) {
	m.Provider = val
}

// GetVersion returns the Version property
func (m GatewayInfo) GetVersion() *VersionInfo {
	return m.Version
}

// SetVersion sets the Version property
func (m *GatewayInfo) SetVersion(val *VersionInfo) {
	m.Version = val
}
