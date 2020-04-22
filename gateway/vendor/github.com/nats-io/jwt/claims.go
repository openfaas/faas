/*
 * Copyright 2018-2019 The NATS Authors
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
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nkeys"
)

// ClaimType is used to indicate the type of JWT being stored in a Claim
type ClaimType string

const (
	// AccountClaim is the type of an Account JWT
	AccountClaim = "account"
	//ActivationClaim is the type of an activation JWT
	ActivationClaim = "activation"
	//UserClaim is the type of an user JWT
	UserClaim = "user"
	//ServerClaim is the type of an server JWT
	ServerClaim = "server"
	//ClusterClaim is the type of an cluster JWT
	ClusterClaim = "cluster"
	//OperatorClaim is the type of an operator JWT
	OperatorClaim = "operator"
)

// Claims is a JWT claims
type Claims interface {
	Claims() *ClaimsData
	Encode(kp nkeys.KeyPair) (string, error)
	ExpectedPrefixes() []nkeys.PrefixByte
	Payload() interface{}
	String() string
	Validate(vr *ValidationResults)
	Verify(payload string, sig []byte) bool
}

// ClaimsData is the base struct for all claims
type ClaimsData struct {
	Audience  string    `json:"aud,omitempty"`
	Expires   int64     `json:"exp,omitempty"`
	ID        string    `json:"jti,omitempty"`
	IssuedAt  int64     `json:"iat,omitempty"`
	Issuer    string    `json:"iss,omitempty"`
	Name      string    `json:"name,omitempty"`
	NotBefore int64     `json:"nbf,omitempty"`
	Subject   string    `json:"sub,omitempty"`
	Tags      TagList   `json:"tags,omitempty"`
	Type      ClaimType `json:"type,omitempty"`
}

// Prefix holds the prefix byte for an NKey
type Prefix struct {
	nkeys.PrefixByte
}

func encodeToString(d []byte) string {
	return base64.RawURLEncoding.EncodeToString(d)
}

func decodeString(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func serialize(v interface{}) (string, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return encodeToString(j), nil
}

func (c *ClaimsData) doEncode(header *Header, kp nkeys.KeyPair, claim Claims) (string, error) {
	if header == nil {
		return "", errors.New("header is required")
	}

	if kp == nil {
		return "", errors.New("keypair is required")
	}

	if c.Subject == "" {
		return "", errors.New("subject is not set")
	}

	h, err := serialize(header)
	if err != nil {
		return "", err
	}

	issuerBytes, err := kp.PublicKey()
	if err != nil {
		return "", err
	}

	prefixes := claim.ExpectedPrefixes()
	if prefixes != nil {
		ok := false
		for _, p := range prefixes {
			switch p {
			case nkeys.PrefixByteAccount:
				if nkeys.IsValidPublicAccountKey(issuerBytes) {
					ok = true
				}
			case nkeys.PrefixByteOperator:
				if nkeys.IsValidPublicOperatorKey(issuerBytes) {
					ok = true
				}
			case nkeys.PrefixByteServer:
				if nkeys.IsValidPublicServerKey(issuerBytes) {
					ok = true
				}
			case nkeys.PrefixByteCluster:
				if nkeys.IsValidPublicClusterKey(issuerBytes) {
					ok = true
				}
			case nkeys.PrefixByteUser:
				if nkeys.IsValidPublicUserKey(issuerBytes) {
					ok = true
				}
			}
		}
		if !ok {
			return "", fmt.Errorf("unable to validate expected prefixes - %v", prefixes)
		}
	}

	c.Issuer = string(issuerBytes)
	c.IssuedAt = time.Now().UTC().Unix()

	c.ID, err = c.hash()
	if err != nil {
		return "", err
	}

	payload, err := serialize(claim)
	if err != nil {
		return "", err
	}

	sig, err := kp.Sign([]byte(payload))
	if err != nil {
		return "", err
	}
	eSig := encodeToString(sig)
	return fmt.Sprintf("%s.%s.%s", h, payload, eSig), nil
}

func (c *ClaimsData) hash() (string, error) {
	j, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	h := sha512.New512_256()
	h.Write(j)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h.Sum(nil)), nil
}

// Encode encodes a claim into a JWT token. The claim is signed with the
// provided nkey's private key
func (c *ClaimsData) Encode(kp nkeys.KeyPair, payload Claims) (string, error) {
	return c.doEncode(&Header{TokenTypeJwt, AlgorithmNkey}, kp, payload)
}

// Returns a JSON representation of the claim
func (c *ClaimsData) String(claim interface{}) string {
	j, err := json.MarshalIndent(claim, "", "  ")
	if err != nil {
		return ""
	}
	return string(j)
}

func parseClaims(s string, target Claims) error {
	h, err := decodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(h, &target)
}

// Verify verifies that the encoded payload was signed by the
// provided public key. Verify is called automatically with
// the claims portion of the token and the public key in the claim.
// Client code need to insure that the public key in the
// claim is trusted.
func (c *ClaimsData) Verify(payload string, sig []byte) bool {
	// decode the public key
	kp, err := nkeys.FromPublicKey(c.Issuer)
	if err != nil {
		return false
	}
	if err := kp.Verify([]byte(payload), sig); err != nil {
		return false
	}
	return true
}

// Validate checks a claim to make sure it is valid. Validity checks
// include expiration and not before constraints.
func (c *ClaimsData) Validate(vr *ValidationResults) {
	now := time.Now().UTC().Unix()
	if c.Expires > 0 && now > c.Expires {
		vr.AddTimeCheck("claim is expired")
	}

	if c.NotBefore > 0 && c.NotBefore > now {
		vr.AddTimeCheck("claim is not yet valid")
	}
}

// IsSelfSigned returns true if the claims issuer is the subject
func (c *ClaimsData) IsSelfSigned() bool {
	return c.Issuer == c.Subject
}

// Decode takes a JWT string decodes it and validates it
// and return the embedded Claims. If the token header
// doesn't match the expected algorithm, or the claim is
// not valid or verification fails an error is returned.
func Decode(token string, target Claims) error {
	// must have 3 chunks
	chunks := strings.Split(token, ".")
	if len(chunks) != 3 {
		return errors.New("expected 3 chunks")
	}

	_, err := parseHeaders(chunks[0])
	if err != nil {
		return err
	}

	if err := parseClaims(chunks[1], target); err != nil {
		return err
	}

	sig, err := decodeString(chunks[2])
	if err != nil {
		return err
	}

	if !target.Verify(chunks[1], sig) {
		return errors.New("claim failed signature verification")
	}

	prefixes := target.ExpectedPrefixes()
	if prefixes != nil {
		ok := false
		issuer := target.Claims().Issuer
		for _, p := range prefixes {
			switch p {
			case nkeys.PrefixByteAccount:
				if nkeys.IsValidPublicAccountKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteOperator:
				if nkeys.IsValidPublicOperatorKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteServer:
				if nkeys.IsValidPublicServerKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteCluster:
				if nkeys.IsValidPublicClusterKey(issuer) {
					ok = true
				}
			case nkeys.PrefixByteUser:
				if nkeys.IsValidPublicUserKey(issuer) {
					ok = true
				}
			}
		}
		if !ok {
			return fmt.Errorf("unable to validate expected prefixes - %v", prefixes)
		}
	}

	return nil
}
