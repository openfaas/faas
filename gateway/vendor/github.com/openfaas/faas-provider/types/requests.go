// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

// ScaleServiceRequest scales the service to the requested replica count.
type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
	Namespace   string `json:"namespace,omitempty"`
}

// DeleteFunctionRequest delete a deployed function
type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
	Namespace    string `json:"namespace,omitempty"`
}

// ProviderInfo provides information about the configured provider
type ProviderInfo struct {
	Name          string       `json:"provider"`
	Version       *VersionInfo `json:"version"`
	Orchestration string       `json:"orchestration"`
}

// VersionInfo provides the commit message, sha and release version number
type VersionInfo struct {
	CommitMessage string `json:"commit_message,omitempty"`
	SHA           string `json:"sha"`
	Release       string `json:"release"`
}

// FunctionNamespace is the namespace for a function
type FunctionNamespace struct {
	Name string `json:"name"`

	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}
