package types

import providerTypes "github.com/openfaas/faas-provider/types"

// Platform architecture the gateway is running on
var Arch string

// GatewayInfo provides information about the gateway and it's connected components
type GatewayInfo struct {
	Provider *providerTypes.ProviderInfo `json:"provider"`
	Version  *providerTypes.VersionInfo  `json:"version"`
	Arch     string                      `json:"arch"`
}
