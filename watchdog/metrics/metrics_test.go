package metrics

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func Test_Register_ProvidesBytes(t *testing.T) {

	metricsPort := 31111

	metricsServer := MetricsServer{}
	metricsServer.Register(metricsPort)

	cancel := make(chan bool)
	go metricsServer.Serve(cancel)

	defer func() {
		cancel <- true
	}()

	retries := 10

	for i := 0; i < retries; i++ {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/metrics", metricsPort), nil)

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			t.Logf("cannot get metrics, or not ready: %s", err.Error())

			time.Sleep(time.Millisecond * 100)
			continue
		}

		wantStatus := http.StatusOK
		if res.StatusCode != wantStatus {
			t.Errorf("metrics gave wrong status, want: %d, got: %d", wantStatus, res.StatusCode)
			t.Fail()
			return
		}

		if res.Body == nil {
			t.Errorf("metrics response should have a body")
			t.Fail()
			return
		}
		defer res.Body.Close()

		return
	}

	t.Errorf("unable to get expected response from metrics server")
	t.Fail()
}
