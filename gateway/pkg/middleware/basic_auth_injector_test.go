// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
