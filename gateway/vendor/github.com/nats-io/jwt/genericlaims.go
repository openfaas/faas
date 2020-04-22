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

import "github.com/nats-io/nkeys"

// GenericClaims can be used to read a JWT as a map for any non-generic fields
type GenericClaims struct {
	ClaimsData
	Data map[string]interface{} `json:"nats,omitempty"`
}

// NewGenericClaims creates a map-based Claims
func NewGenericClaims(subject string) *GenericClaims {
	if subject == "" {
		return nil
	}
	c := GenericClaims{}
	c.Subject = subject
	c.Data = make(map[string]interface{})
	return &c
}

// DecodeGeneric takes a JWT string and decodes it into a ClaimsData and map
func DecodeGeneric(token string) (*GenericClaims, error) {
	v := GenericClaims{}
	if err := Decode(token, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Claims returns the standard part of the generic claim
func (gc *GenericClaims) Claims() *ClaimsData {
	return &gc.ClaimsData
}

// Payload returns the custom part of the claims data
func (gc *GenericClaims) Payload() interface{} {
	return &gc.Data
}

// Encode takes a generic claims and creates a JWT string
func (gc *GenericClaims) Encode(pair nkeys.KeyPair) (string, error) {
	return gc.ClaimsData.Encode(pair, gc)
}

// Validate checks the generic part of the claims data
func (gc *GenericClaims) Validate(vr *ValidationResults) {
	gc.ClaimsData.Validate(vr)
}

func (gc *GenericClaims) String() string {
	return gc.ClaimsData.String(gc)
}

// ExpectedPrefixes returns the types allowed to encode a generic JWT, which is nil for all
func (gc *GenericClaims) ExpectedPrefixes() []nkeys.PrefixByte {
	return nil
}
