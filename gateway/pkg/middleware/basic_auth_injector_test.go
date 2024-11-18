// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) OpenFaaS Author(s). All rights reserved.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Inject_WithNilRequestAndNilCredentials(t *testing.T) {
	injector := BasicAuthInjector{}
	injector.Inject(nil)
}

func Test_Inject_WithRequestButNilCredentials(t *testing.T) {
	injector := BasicAuthInjector{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	injector.Inject(req)
}
