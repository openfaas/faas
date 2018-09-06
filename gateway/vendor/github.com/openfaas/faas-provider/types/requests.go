// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
}

// InfoRequest provides information about the underlying provider
type InfoRequest struct {
	Provider      string          `json:"provider"`
	Version       ProviderVersion `json:"version"`
	Orchestration string          `json:"orchestration"`
}

// ProviderVersion provides the commit sha and release version number of the underlying provider
type ProviderVersion struct {
	SHA     string `json:"sha"`
	Release string `json:"release"`
}
