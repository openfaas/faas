package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

// MakeExternalAuthHandler make an authentication proxy handler
func MakeExternalAuthHandler(next http.HandlerFunc, upstreamTimeout time.Duration, upstreamURL string, passBody bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(http.MethodGet, upstreamURL, nil)

		copyHeaders(req.Header, &r.Header)

		deadlineContext, cancel := context.WithTimeout(
			context.Background(),
			upstreamTimeout)

		defer cancel()

		res, err := http.DefaultClient.Do(req.WithContext(deadlineContext))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("ExternalAuthHandler: %s", err.Error())
			return
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		if res.StatusCode == http.StatusOK {
			next.ServeHTTP(w, r)
			return
		}

		copyHeaders(w.Header(), &res.Header)
		w.WriteHeader(res.StatusCode)

		if res.Body != nil {
			io.Copy(w, res.Body)
		}
	}
}
