// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package auth

import (
	"crypto/subtle"
	"net/http"
)

// DecorateWithBasicAuth enforces basic auth as a middleware with given credentials
func DecorateWithBasicAuth(next http.HandlerFunc, credentials *BasicAuthCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, password, ok := r.BasicAuth()

		const noMatch = 0
		if !ok ||
			user != credentials.User ||
			subtle.ConstantTimeCompare([]byte(credentials.Password), []byte(password)) == noMatch {

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	}
}
