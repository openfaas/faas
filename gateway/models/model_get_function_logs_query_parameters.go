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

// GetFunctionLogsQueryParameters is an object.
type GetFunctionLogsQueryParameters struct {
	// Name: Function name
	Name string `json:"name" mapstructure:"name"`
	// Namespace: Namespace of the function
	Namespace string `json:"namespace,omitempty" mapstructure:"namespace,omitempty"`
	// Instance: Instance of the function
	Instance string `json:"instance,omitempty" mapstructure:"instance,omitempty"`
	// Tail: Sets the maximum number of log messages to return, <=0 means unlimited
	Tail int32 `json:"tail,omitempty" mapstructure:"tail,omitempty"`
	// Follow: When true, the request will stream logs until the request timeout
	Follow bool `json:"follow,omitempty" mapstructure:"follow,omitempty"`
	// Since: Only return logs after a specific date (RFC3339)
	Since time.Time `json:"since" mapstructure:"since"`
}

// Validate implements basic validation for this model
func (m GetFunctionLogsQueryParameters) Validate() error {
	return validation.Errors{}.Filter()
}

// GetName returns the Name property
func (m GetFunctionLogsQueryParameters) GetName() string {
	return m.Name
}

// SetName sets the Name property
func (m *GetFunctionLogsQueryParameters) SetName(val string) {
	m.Name = val
}

// GetNamespace returns the Namespace property
func (m GetFunctionLogsQueryParameters) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace property
func (m *GetFunctionLogsQueryParameters) SetNamespace(val string) {
	m.Namespace = val
}

// GetInstance returns the Instance property
func (m GetFunctionLogsQueryParameters) GetInstance() string {
	return m.Instance
}

// SetInstance sets the Instance property
func (m *GetFunctionLogsQueryParameters) SetInstance(val string) {
	m.Instance = val
}

// GetTail returns the Tail property
func (m GetFunctionLogsQueryParameters) GetTail() int32 {
	return m.Tail
}

// SetTail sets the Tail property
func (m *GetFunctionLogsQueryParameters) SetTail(val int32) {
	m.Tail = val
}

// GetFollow returns the Follow property
func (m GetFunctionLogsQueryParameters) GetFollow() bool {
	return m.Follow
}

// SetFollow sets the Follow property
func (m *GetFunctionLogsQueryParameters) SetFollow(val bool) {
	m.Follow = val
}

// GetSince returns the Since property
func (m GetFunctionLogsQueryParameters) GetSince() time.Time {
	return m.Since
}

// SetSince sets the Since property
func (m *GetFunctionLogsQueryParameters) SetSince(val time.Time) {
	m.Since = val
}
