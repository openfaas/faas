// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package requests

import (
	"encoding/json"
	"testing"
)

// TestUnmarshallAlert is an exploratory test from TDD'ing the struct to parse a Prometheus alert
func TestUnmarshallAlert(t *testing.T) {
	file := []byte(`{
    "receiver": "scale-up",
    "status": "firing",
    "alerts": [{
        "status": "firing",
        "labels": {
            "alertname": "APIHighInvocationRate",
            "code": "200",
            "function_name": "func_nodeinfo",
            "instance": "gateway:8080",
            "job": "gateway",
            "monitor": "faas-monitor",
            "service": "gateway",
            "severity": "major",
            "value": "8.998200359928017"
        },
        "annotations": {
            "description": "High invocation total on gateway:8080",
            "summary": "High invocation total on gateway:8080"
        },
        "startsAt": "2017-03-15T15:52:57.805Z",
        "endsAt": "0001-01-01T00:00:00Z",
        "generatorURL": "http://4156cb797423:9090/graph?g0.expr=rate%28gateway_function_invocation_total%5B10s%5D%29+%3E+5\u0026g0.tab=0"
    }],
    "groupLabels": {
        "alertname": "APIHighInvocationRate",
        "service": "gateway"
    },
    "commonLabels": {
        "alertname": "APIHighInvocationRate",
        "code": "200",
        "function_name": "func_nodeinfo",
        "instance": "gateway:8080",
        "job": "gateway",
        "monitor": "faas-monitor",
        "service": "gateway",
        "severity": "major",
        "value": "8.998200359928017"
    },
    "commonAnnotations": {
        "description": "High invocation total on gateway:8080",
        "summary": "High invocation total on gateway:8080"
    },
    "externalURL": "http://f054879d97db:9093",
    "version": "3",
    "groupKey": 18195285354214864953
}`)

	var alert PrometheusAlert
	err := json.Unmarshal(file, &alert)

	if err != nil {
		t.Fatal(err)
	}

	if (len(alert.Status)) == 0 {
		t.Fatal("No status read")
	}

	if (len(alert.Receiver)) == 0 {
		t.Fatal("No status read")
	}

	if (len(alert.Alerts)) == 0 {
		t.Fatal("No alerts read")
	}

	if (len(alert.Alerts[0].Labels.AlertName)) == 0 {
		t.Fatal("No alerts name")
	}

	if (len(alert.Alerts[0].Labels.FunctionName)) == 0 {
		t.Fatal("No function name read")
	}

}
