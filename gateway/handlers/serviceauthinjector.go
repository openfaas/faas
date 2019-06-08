package handlers

import "net/http"

type AuthInjector interface {
	Inject(r *http.Request)
}
