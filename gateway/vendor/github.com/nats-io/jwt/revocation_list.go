/*
 * Copyright 2020 The NATS Authors
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
	"time"
)

const All = "*"

// RevocationList is used to store a mapping of public keys to unix timestamps
type RevocationList map[string]int64

// Revoke enters a revocation by publickey and timestamp into this export
// If there is already a revocation for this public key that is newer, it is kept.
func (r RevocationList) Revoke(pubKey string, timestamp time.Time) {
	newTS := timestamp.Unix()
	if ts, ok := r[pubKey]; ok && ts > newTS {
		return
	}

	r[pubKey] = newTS
}

// ClearRevocation removes any revocation for the public key
func (r RevocationList) ClearRevocation(pubKey string) {
	delete(r, pubKey)
}

// IsRevoked checks if the public key is in the revoked list with a timestamp later than
// the one passed in. Generally this method is called with an issue time but other time's can
// be used for testing.
func (r RevocationList) IsRevoked(pubKey string, timestamp time.Time) bool {
	if r.allRevoked(timestamp) {
		return true
	}
	ts, ok := r[pubKey]
	return ok && ts >= timestamp.Unix()
}

// allRevoked returns true if All is set and the timestamp is later or same as the
// one passed. This is called by IsRevoked.
func (r RevocationList) allRevoked(timestamp time.Time) bool {
	ts, ok := r[All]
	return ok && ts >= timestamp.Unix()
}
