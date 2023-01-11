package middleware

import "net/http"

// AuthInjector is an interface for injecting authentication information into a request
// which will be proxied or made to a remote/upstream service.
type AuthInjector interface {
	Inject(r *http.Request)
}
