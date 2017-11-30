package handlers

import "net/http"

type CorsHandler struct {
	Upstream    *http.Handler
	AllowedHost string
}

func (c CorsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// https://raw.githubusercontent.com/openfaas/store/master/store.json
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Origin", c.AllowedHost)

	(*c.Upstream).ServeHTTP(w, r)
}

func DecorateWithCORS(upstream http.Handler, allowedHost string) http.Handler {
	return CorsHandler{
		Upstream:    &upstream,
		AllowedHost: allowedHost,
	}
}
