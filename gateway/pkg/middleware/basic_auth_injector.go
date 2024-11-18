// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s). All rights reserved.

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
