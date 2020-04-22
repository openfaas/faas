/*
 * Copyright 2018 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jwt

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nats-io/nkeys"
)

// Operator specific claims
type Operator struct {
	// Slice of real identies (like websites) that can be used to identify the operator.
	Identities []Identity `json:"identity,omitempty"`
	// Slice of other operator NKeys that can be used to sign on behalf of the main
	// operator identity.
	SigningKeys StringList `json:"signing_keys,omitempty"`
	// AccountServerURL is a partial URL like "https://host.domain.org:<port>/jwt/v1"
	// tools will use the prefix and build queries by appending /accounts/<account_id>
	// or /operator to the path provided. Note this assumes that the account server
	// can handle requests in a nats-account-server compatible way. See
	// https://github.com/nats-io/nats-account-server.
	AccountServerURL string `json:"account_server_url,omitempty"`
	// A list of NATS urls (tls://host:port) where tools can connect to the server
	// using proper credentials.
	OperatorServiceURLs StringList `json:"operator_service_urls,omitempty"`
}

// Validate checks the validity of the operators contents
func (o *Operator) Validate(vr *ValidationResults) {
	if err := o.validateAccountServerURL(); err != nil {
		vr.AddError(err.Error())
	}

	for _, v := range o.validateOperatorServiceURLs() {
		if v != nil {
			vr.AddError(v.Error())
		}
	}

	for _, i := range o.Identities {
		i.Validate(vr)
	}

	for _, k := range o.SigningKeys {
		if !nkeys.IsValidPublicOperatorKey(k) {
			vr.AddError("%s is not an operator public key", k)
		}
	}
}

func (o *Operator) validateAccountServerURL() error {
	if o.AccountServerURL != "" {
		// We don't care what kind of URL it is so long as it parses
		// and has a protocol. The account server may impose additional
		// constraints on the type of URLs that it is able to notify to
		u, err := url.Parse(o.AccountServerURL)
		if err != nil {
			return fmt.Errorf("error parsing account server url: %v", err)
		}
		if u.Scheme == "" {
			return fmt.Errorf("account server url %q requires a protocol", o.AccountServerURL)
		}
	}
	return nil
}

// ValidateOperatorServiceURL returns an error if the URL is not a valid NATS or TLS url.
func ValidateOperatorServiceURL(v string) error {
	// should be possible for the service url to not be expressed
	if v == "" {
		return nil
	}
	u, err := url.Parse(v)
	if err != nil {
		return fmt.Errorf("error parsing operator service url %q: %v", v, err)
	}

	if u.User != nil {
		return fmt.Errorf("operator service url %q - credentials are not supported", v)
	}

	if u.Path != "" {
		return fmt.Errorf("operator service url %q - paths are not supported", v)
	}

	lcs := strings.ToLower(u.Scheme)
	switch lcs {
	case "nats":
		return nil
	case "tls":
		return nil
	default:
		return fmt.Errorf("operator service url %q - protocol not supported (only 'nats' or 'tls' only)", v)
	}
}

func (o *Operator) validateOperatorServiceURLs() []error {
	var errors []error
	for _, v := range o.OperatorServiceURLs {
		if v != "" {
			if err := ValidateOperatorServiceURL(v); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

// OperatorClaims define the data for an operator JWT
type OperatorClaims struct {
	ClaimsData
	Operator `json:"nats,omitempty"`
}

// NewOperatorClaims creates a new operator claim with the specified subject, which should be an operator public key
func NewOperatorClaims(subject string) *OperatorClaims {
	if subject == "" {
		return nil
	}
	c := &OperatorClaims{}
	c.Subject = subject
	return c
}

// DidSign checks the claims against the operator's public key and its signing keys
func (oc *OperatorClaims) DidSign(op Claims) bool {
	if op == nil {
		return false
	}
	issuer := op.Claims().Issuer
	if issuer == oc.Subject {
		return true
	}
	return oc.SigningKeys.Contains(issuer)
}

// Deprecated: AddSigningKey, use claim.SigningKeys.Add()
func (oc *OperatorClaims) AddSigningKey(pk string) {
	oc.SigningKeys.Add(pk)
}

// Encode the claims into a JWT string
func (oc *OperatorClaims) Encode(pair nkeys.KeyPair) (string, error) {
	if !nkeys.IsValidPublicOperatorKey(oc.Subject) {
		return "", errors.New("expected subject to be an operator public key")
	}
	err := oc.validateAccountServerURL()
	if err != nil {
		return "", err
	}
	oc.ClaimsData.Type = OperatorClaim
	return oc.ClaimsData.Encode(pair, oc)
}

// DecodeOperatorClaims tries to create an operator claims from a JWt string
func DecodeOperatorClaims(token string) (*OperatorClaims, error) {
	v := OperatorClaims{}
	if err := Decode(token, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (oc *OperatorClaims) String() string {
	return oc.ClaimsData.String(oc)
}

// Payload returns the operator specific data for an operator JWT
func (oc *OperatorClaims) Payload() interface{} {
	return &oc.Operator
}

// Validate the contents of the claims
func (oc *OperatorClaims) Validate(vr *ValidationResults) {
	oc.ClaimsData.Validate(vr)
	oc.Operator.Validate(vr)
}

// ExpectedPrefixes defines the nkey types that can sign operator claims, operator
func (oc *OperatorClaims) ExpectedPrefixes() []nkeys.PrefixByte {
	return []nkeys.PrefixByte{nkeys.PrefixByteOperator}
}

// Claims returns the generic claims data
func (oc *OperatorClaims) Claims() *ClaimsData {
	return &oc.ClaimsData
}
