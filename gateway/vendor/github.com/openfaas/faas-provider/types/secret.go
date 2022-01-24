package types

// Secret for underlying orchestrator
type Secret struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace if applicable for the secret
	Namespace string `json:"namespace,omitempty"`

	// Value is a string representing the string's value
	Value string `json:"value,omitempty"`

	// RawValue can be used to provide binary data when
	// Value is not set
	RawValue []byte `json:"rawValue,omitempty"`
}
