package requests

import (
	"fmt"
	"testing"

	"io/ioutil"

	"encoding/json"

	"github.com/alexellis/faas/gateway/requests"
)

func TestUnmarshallAlert(t *testing.T) {
	file, _ := ioutil.ReadFile("./test_alert.json")

	var alert requests.PrometheusAlert
	err := json.Unmarshal(file, &alert)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("OK", string(file), alert)
	if (len(alert.Status)) == 0 {
		t.Fatal("No status read")
	}
	if (len(alert.Receiver)) == 0 {
		t.Fatal("No status read")
	}
}
