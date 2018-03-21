package handlers

import "net/http"

// CORSHandler set custom CORS instructions for the store.
type CORSHandler struct {
	Upstream    *http.Handler
	AllowedHost string
}

func (c CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// https://raw.githubusercontent.com/openfaas/store/master/store.json
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
	w.Header().Set("Access-Control-Allow-Origin", c.AllowedHost)

	(*c.Upstream).ServeHTTP(w, r)
}

// DecorateWithCORS decorate a handler with CORS-injecting middleware
func DecorateWithCORS(upstream http.Handler, allowedHost string) http.Handler {
	return CORSHandler{
		Upstream:    &upstream,
		AllowedHost: allowedHost,
	}
}
