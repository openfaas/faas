package metrics

import (
	"testing"

	"github.com/openfaas/faas/gateway/requests"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type metricResult struct {
	value  float64
	labels map[string]string
}

func labels2Map(labels []*dto.LabelPair) map[string]string {
	res := map[string]string{}
	for _, l := range labels {
		res[l.GetName()] = l.GetValue()
	}
	return res
}

func readGauge(g prometheus.Metric) metricResult {
	m := &dto.Metric{}
	g.Write(m)

	return metricResult{
		value:  m.GetGauge().GetValue(),
		labels: labels2Map(m.GetLabel()),
	}
}

func Test_Describe_DescribesThePrometheusMetrics(t *testing.T) {
	metricsOptions := BuildMetricsOptions()
	exporter := NewExporter(metricsOptions, nil)

	ch := make(chan *prometheus.Desc)
	defer close(ch)

	go exporter.Describe(ch)

	d := <-ch
	expectedGatewayFunctionInvocationDesc := `Desc{fqName: "gateway_function_invocation_total", help: "Individual function metrics", constLabels: {}, variableLabels: [function_name code]}`
	actualGatewayFunctionInvocationDesc := d.String()
	if expectedGatewayFunctionInvocationDesc != actualGatewayFunctionInvocationDesc {
		t.Errorf("Want %s, got: %s", expectedGatewayFunctionInvocationDesc, actualGatewayFunctionInvocationDesc)
	}

	d = <-ch
	expectedGatewayFunctionsHistogramDesc := `Desc{fqName: "gateway_functions_seconds", help: "Function time taken", constLabels: {}, variableLabels: [function_name]}`
	actualGatewayFunctionsHistogramDesc := d.String()
	if expectedGatewayFunctionsHistogramDesc != actualGatewayFunctionsHistogramDesc {
		t.Errorf("Want %s, got: %s", expectedGatewayFunctionsHistogramDesc, actualGatewayFunctionsHistogramDesc)
	}

	d = <-ch
	expectedServiceReplicasGaugeDesc := `Desc{fqName: "gateway_service_count", help: "Docker service replicas", constLabels: {}, variableLabels: [function_name]}`
	actualServiceReplicasGaugeDesc := d.String()
	if expectedServiceReplicasGaugeDesc != actualServiceReplicasGaugeDesc {
		t.Errorf("Want %s, got: %s", expectedServiceReplicasGaugeDesc, actualServiceReplicasGaugeDesc)
	}

}

func Test_Collect_CollectsTheNumberOfReplicasOfAService(t *testing.T) {
	metricsOptions := BuildMetricsOptions()
	exporter := NewExporter(metricsOptions, nil)

	expectedService := requests.Function{
		Name:     "function_with_two_replica",
		Replicas: 2,
	}

	exporter.services = []requests.Function{expectedService}

	ch := make(chan prometheus.Metric)
	defer close(ch)

	go exporter.Collect(ch)

	g := (<-ch).(prometheus.Gauge)
	result := readGauge(g)
	if expectedService.Name != result.labels["function_name"] {
		t.Errorf("Want %s, got %s", expectedService.Name, result.labels["function_name"])
	}
	expectedReplicas := float64(expectedService.Replicas)
	if expectedReplicas != result.value {
		t.Errorf("Want %f, got %f", expectedReplicas, result.value)
	}
}
