package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/alexellis/faas/gateway/metrics"
	"github.com/alexellis/faas/gateway/queue"
	"github.com/gorilla/mux"
)

// MakeQueuedProxy accepts work onto a queue
func MakeQueuedProxy(metrics metrics.MetricOptions, wildcard bool, logger *logrus.Logger, canQueueRequests queue.CanQueueRequests) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		vars := mux.Vars(r)
		name := vars["name"]
		req := &queue.Request{
			Function:    name,
			Body:        body,
			Method:      r.Method,
			QueryString: r.URL.RawQuery,
			Header:      r.Header,
		}

		err = canQueueRequests.Queue(req)
		if err != nil {
			fmt.Println(err)
		}
	}
}
