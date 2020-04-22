package jwt

import (
	"time"
)

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
// the one passed in. Generally this method is called with time.Now() but other time's can
// be used for testing.
func (r RevocationList) IsRevoked(pubKey string, timestamp time.Time) bool {
	ts, ok := r[pubKey]
	return ok && ts > timestamp.Unix()
}
