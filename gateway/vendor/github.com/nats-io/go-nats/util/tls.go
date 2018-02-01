// Copyright 2017 Apcera Inc. All rights reserved.
// +build go1.8

package util

import "crypto/tls"

// CloneTLSConfig returns a copy of c.
func CloneTLSConfig(c *tls.Config) *tls.Config {
	if c == nil {
		return &tls.Config{}
	}

	return c.Clone()
}
