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
	"errors"

	"github.com/nats-io/nkeys"
)

// User defines the user specific data in a user JWT
type User struct {
	Permissions
	Limits
	BearerToken bool `json:"bearer_token,omitempty"`
}

// Validate checks the permissions and limits in a User jwt
func (u *User) Validate(vr *ValidationResults) {
	u.Permissions.Validate(vr)
	u.Limits.Validate(vr)
	// When BearerToken is true server will ignore any nonce-signing verification
}

// UserClaims defines a user JWT
type UserClaims struct {
	ClaimsData
	User `json:"nats,omitempty"`
	// IssuerAccount stores the public key for the account the issuer represents.
	// When set, the claim was issued by a signing key.
	IssuerAccount string `json:"issuer_account,omitempty"`
}

// NewUserClaims creates a user JWT with the specific subject/public key
func NewUserClaims(subject string) *UserClaims {
	if subject == "" {
		return nil
	}
	c := &UserClaims{}
	c.Subject = subject
	return c
}

// Encode tries to turn the user claims into a JWT string
func (u *UserClaims) Encode(pair nkeys.KeyPair) (string, error) {
	if !nkeys.IsValidPublicUserKey(u.Subject) {
		return "", errors.New("expected subject to be user public key")
	}
	u.ClaimsData.Type = UserClaim
	return u.ClaimsData.Encode(pair, u)
}

// DecodeUserClaims tries to parse a user claims from a JWT string
func DecodeUserClaims(token string) (*UserClaims, error) {
	v := UserClaims{}
	if err := Decode(token, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Validate checks the generic and specific parts of the user jwt
func (u *UserClaims) Validate(vr *ValidationResults) {
	u.ClaimsData.Validate(vr)
	u.User.Validate(vr)
	if u.IssuerAccount != "" && !nkeys.IsValidPublicAccountKey(u.IssuerAccount) {
		vr.AddError("account_id is not an account public key")
	}
}

// ExpectedPrefixes defines the types that can encode a user JWT, account
func (u *UserClaims) ExpectedPrefixes() []nkeys.PrefixByte {
	return []nkeys.PrefixByte{nkeys.PrefixByteAccount}
}

// Claims returns the generic data from a user jwt
func (u *UserClaims) Claims() *ClaimsData {
	return &u.ClaimsData
}

// Payload returns the user specific data from a user JWT
func (u *UserClaims) Payload() interface{} {
	return &u.User
}

func (u *UserClaims) String() string {
	return u.ClaimsData.String(u)
}

// IsBearerToken returns true if nonce-signing requirements should be skipped
func (u *UserClaims) IsBearerToken() bool {
	return u.BearerToken
}
