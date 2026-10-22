package metrics

import (
	"testing"

	types "github.com/openfaas/faas-provider/types"
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
	exporter := NewExporter(metricsOptions, nil, "openfaas-fn")

	ch := make(chan *prometheus.Desc)
	// defer close(ch)

	go exporter.Describe(ch)

	d := <-ch
	expectedGatewayFunctionInvocationDesc := prometheus.NewDesc("gateway_function_invocation_total", "Function metrics", []string{"function_name", "code"}, nil).String()
	actualGatewayFunctionInvocationDesc := d.String()
	if expectedGatewayFunctionInvocationDesc != actualGatewayFunctionInvocationDesc {
		t.Errorf("Want\n%s\ngot\n%s", expectedGatewayFunctionInvocationDesc, actualGatewayFunctionInvocationDesc)
	}

	d = <-ch
	expectedGatewayFunctionsHistogramDesc := prometheus.NewDesc("gateway_functions_seconds", "Function time taken", []string{"function_name", "code"}, nil).String()
	actualGatewayFunctionsHistogramDesc := d.String()
	if expectedGatewayFunctionsHistogramDesc != actualGatewayFunctionsHistogramDesc {
		t.Errorf("Want\n%s\ngot\n%s", expectedGatewayFunctionsHistogramDesc, actualGatewayFunctionsHistogramDesc)
	}

	d = <-ch
	expectedServiceReplicasGaugeDesc := prometheus.NewDesc("gateway_service_count", "Current count of replicas for function", []string{"function_name"}, nil).String()
	actualServiceReplicasGaugeDesc := d.String()
	if expectedServiceReplicasGaugeDesc != actualServiceReplicasGaugeDesc {
		t.Errorf("Want\n%s\ngot\n%s", expectedServiceReplicasGaugeDesc, actualServiceReplicasGaugeDesc)
	}

}

func Test_Collect_CollectsTheNumberOfReplicasOfAService(t *testing.T) {
	metricsOptions := BuildMetricsOptions()
	exporter := NewExporter(metricsOptions, nil, "openfaas-fn")

	expectedService := types.FunctionStatus{
		Name:     "function_with_two_replica",
		Replicas: 2,
	}

	exporter.services = []types.FunctionStatus{expectedService}

	ch := make(chan prometheus.Metric)

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
	ch = nil

}
