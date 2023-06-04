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

// VersionInfo is an object.
type VersionInfo struct {
	// CommitMessage:
	CommitMessage string `json:"commit_message,omitempty" mapstructure:"commit_message,omitempty"`
	// Release:
	Release string `json:"release" mapstructure:"release"`
	// Sha:
	Sha string `json:"sha" mapstructure:"sha"`
}

// Validate implements basic validation for this model
func (m VersionInfo) Validate() error {
	return validation.Errors{}.Filter()
}

// GetCommitMessage returns the CommitMessage property
func (m VersionInfo) GetCommitMessage() string {
	return m.CommitMessage
}

// SetCommitMessage sets the CommitMessage property
func (m *VersionInfo) SetCommitMessage(val string) {
	m.CommitMessage = val
}

// GetRelease returns the Release property
func (m VersionInfo) GetRelease() string {
	return m.Release
}

// SetRelease sets the Release property
func (m *VersionInfo) SetRelease(val string) {
	m.Release = val
}

// GetSha returns the Sha property
func (m VersionInfo) GetSha() string {
	return m.Sha
}

// SetSha sets the Sha property
func (m *VersionInfo) SetSha(val string) {
	m.Sha = val
}
