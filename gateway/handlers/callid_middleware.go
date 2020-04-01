// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/distribution/uuid"
)

// MakeCallIDMiddleware middleware tags a request with a uid
func MakeCallIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if len(r.Header.Get("X-Call-Id")) == 0 {
			callID := uuid.Generate().String()
			r.Header.Add("X-Call-Id", callID)
			w.Header().Add("X-Call-Id", callID)
		}

		r.Header.Add("X-Start-Time", fmt.Sprintf("%d", start.UTC().UnixNano()))
		w.Header().Add("X-Start-Time", fmt.Sprintf("%d", start.UTC().UnixNano()))

		printContextStatus(r.Context())

		next(w, r)

		printContextStatus(r.Context())
	}
}

func printContextStatus(ctx context.Context) {
	select {
	case <-ctx.Done():
		fmt.Println("Context closed")
		fmt.Println(ctx.Err())
	default:
		fmt.Println("Context is still valid")
	}
}
