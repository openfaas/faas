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

	"github.com/nats-io/nkeys"
)

// Server defines the custom part of a server jwt
type Server struct {
	Permissions
	Cluster string `json:"cluster,omitempty"`
}

// Validate checks the cluster and permissions for a server JWT
func (s *Server) Validate(vr *ValidationResults) {
	if s.Cluster == "" {
		vr.AddError("servers can't contain an empty cluster")
	}
}

// ServerClaims defines the data in a server JWT
type ServerClaims struct {
	ClaimsData
	Server `json:"nats,omitempty"`
}

// NewServerClaims creates a new server JWT with the specified subject/public key
func NewServerClaims(subject string) *ServerClaims {
	if subject == "" {
		return nil
	}
	c := &ServerClaims{}
	c.Subject = subject
	return c
}

// Encode tries to turn the server claims into a JWT string
func (s *ServerClaims) Encode(pair nkeys.KeyPair) (string, error) {
	if !nkeys.IsValidPublicServerKey(s.Subject) {
		return "", errors.New("expected subject to be a server public key")
	}
	s.ClaimsData.Type = ServerClaim
	return s.ClaimsData.Encode(pair, s)
}

// DecodeServerClaims tries to parse server claims from a JWT string
func DecodeServerClaims(token string) (*ServerClaims, error) {
	v := ServerClaims{}
	if err := Decode(token, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *ServerClaims) String() string {
	return s.ClaimsData.String(s)
}

// Payload returns the server specific data
func (s *ServerClaims) Payload() interface{} {
	return &s.Server
}

// Validate checks the generic and server data in the server claims
func (s *ServerClaims) Validate(vr *ValidationResults) {
	s.ClaimsData.Validate(vr)
	s.Server.Validate(vr)
}

// ExpectedPrefixes defines the types that can encode a server JWT, operator or cluster
func (s *ServerClaims) ExpectedPrefixes() []nkeys.PrefixByte {
	return []nkeys.PrefixByte{nkeys.PrefixByteOperator, nkeys.PrefixByteCluster}
}

// Claims returns the generic data
func (s *ServerClaims) Claims() *ClaimsData {
	return &s.ClaimsData
}
