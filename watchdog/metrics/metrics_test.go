package metrics

import (
	"testing"
	"time"
)

func Test_RegisterServer(t *testing.T) {

	metricsPort := 31111

	metricsServer := MetricsServer{}
	metricsServer.Register(metricsPort)

	cancel := make(chan bool)
	go metricsServer.Serve(cancel)

	time.AfterFunc(time.Millisecond*500, func() {
		cancel <- true
	})
}
