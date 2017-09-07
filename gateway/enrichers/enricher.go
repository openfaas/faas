package enrichers

import "net/http"

// EnricherFunc TODO
type EnricherFunc func(func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)

// EnrichFunc TODO
func EnrichFunc(h func(http.ResponseWriter, *http.Request), enricherfuncs ...EnricherFunc) func(http.ResponseWriter, *http.Request) {
	for _, enricherfunc := range enricherfuncs {
		h = enricherfunc(h)
	}
	return h
}
