package middleware

import "net/http"

type AuthInjector interface {
	Inject(r *http.Request)
}
