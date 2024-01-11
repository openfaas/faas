package types

import "time"

const (
	TypeFunctionUsage = "function_usage"
	TypeAPIAccess     = "api_access"
)

type Event interface {
	EventType() string
}

type FunctionUsageEvent struct {
	Namespace    string        `json:"namespace"`
	FunctionName string        `json:"function_name"`
	Started      time.Time     `json:"started"`
	Duration     time.Duration `json:"duration"`
	MemoryBytes  int64         `json:"memory_bytes"`
}

func (e FunctionUsageEvent) EventType() string {
	return TypeFunctionUsage
}

type APIAccessEvent struct {
	Actor         *Actor    `json:"actor,omitempty"`
	Path          string    `json:"path"`
	Method        string    `json:"method"`
	Actions       []string  `json:"actions"`
	ResponseCode  int       `json:"response_code"`
	CustomMessage string    `json:"custom_message,omitempty"`
	Namespace     string    `json:"namespace,omitempty"`
	Time          time.Time `json:"time"`
}

func (e APIAccessEvent) EventType() string {
	return TypeAPIAccess
}

// Actor is the user that triggered an event.
// Get from OIDC claims, we can add any of the default OIDC profile or email claim fields if desired.
type Actor struct {
	// OIDC subject, a unique identifier of the user.
	Sub string `json:"sub"`
	// Full name of the subject, can be the name of a user of OpenFaaS component.
	Name string `json:"name,omitempty"`
	// OpenFaaS issuer
	Issuer string `json:"issuer,omitempty"`
	// Federated issuer
	FedIssuer string `json:"fed_issuer,omitempty"`
}
