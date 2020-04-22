// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package middleware

import (
	"net/http"

	"github.com/openfaas/faas-provider/auth"
)

type BasicAuthInjector struct {
	Credentials *auth.BasicAuthCredentials
}

func (b BasicAuthInjector) Inject(r *http.Request) {
	if r != nil && b.Credentials != nil {
		r.SetBasicAuth(b.Credentials.User, b.Credentials.Password)
	}
}
