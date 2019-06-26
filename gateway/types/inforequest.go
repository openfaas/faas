package types

// Platform architecture the gateway is running on
var Arch string

// GatewayInfo provides information about the gateway and it's connected components
type GatewayInfo struct {
	Provider *ProviderInfo `json:"provider"`
	Version  *VersionInfo  `json:"version"`
	Arch     string        `json:"arch"`
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
