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

// Cluster stores the cluster specific elements of a cluster JWT
type Cluster struct {
	Trust       []string `json:"identity,omitempty"`
	Accounts    []string `json:"accts,omitempty"`
	AccountURL  string   `json:"accturl,omitempty"`
	OperatorURL string   `json:"opurl,omitempty"`
}

// Validate checks the cluster and permissions for a cluster JWT
func (c *Cluster) Validate(vr *ValidationResults) {
	// fixme validate cluster data
}

// ClusterClaims defines the data in a cluster JWT
type ClusterClaims struct {
	ClaimsData
	Cluster `json:"nats,omitempty"`
}

// NewClusterClaims creates a new cluster JWT with the specified subject/public key
func NewClusterClaims(subject string) *ClusterClaims {
	if subject == "" {
		return nil
	}
	c := &ClusterClaims{}
	c.Subject = subject
	return c
}

// Encode tries to turn the cluster claims into a JWT string
func (c *ClusterClaims) Encode(pair nkeys.KeyPair) (string, error) {
	if !nkeys.IsValidPublicClusterKey(c.Subject) {
		return "", errors.New("expected subject to be a cluster public key")
	}
	c.ClaimsData.Type = ClusterClaim
	return c.ClaimsData.Encode(pair, c)
}

// DecodeClusterClaims tries to parse cluster claims from a JWT string
func DecodeClusterClaims(token string) (*ClusterClaims, error) {
	v := ClusterClaims{}
	if err := Decode(token, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *ClusterClaims) String() string {
	return c.ClaimsData.String(c)
}

// Payload returns the cluster specific data
func (c *ClusterClaims) Payload() interface{} {
	return &c.Cluster
}

// Validate checks the generic and cluster data in the cluster claims
func (c *ClusterClaims) Validate(vr *ValidationResults) {
	c.ClaimsData.Validate(vr)
	c.Cluster.Validate(vr)
}

// ExpectedPrefixes defines the types that can encode a cluster JWT, operator or cluster
func (c *ClusterClaims) ExpectedPrefixes() []nkeys.PrefixByte {
	return []nkeys.PrefixByte{nkeys.PrefixByteOperator, nkeys.PrefixByteCluster}
}

// Claims returns the generic data
func (c *ClusterClaims) Claims() *ClaimsData {
	return &c.ClaimsData
}
