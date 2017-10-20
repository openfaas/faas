package tests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/openfaas/faas/gateway/handlers"
	"github.com/openfaas/faas/gateway/metrics"
	"github.com/openfaas/faas/gateway/requests"
	"golang.org/x/net/context"
)

type testServiceApiClient struct {
	serviceListServices []swarm.Service
	serviceListTasks    []swarm.Task
	serviceListError    error
}

func (c testServiceApiClient) ServiceCreate(ctx context.Context, service swarm.ServiceSpec, options types.ServiceCreateOptions) (types.ServiceCreateResponse, error) {
	return types.ServiceCreateResponse{}, nil
}

func (c testServiceApiClient) ServiceInspectWithRaw(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	return swarm.Service{}, []byte{}, nil
}

func (c testServiceApiClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.serviceListServices, c.serviceListError
}

func (c testServiceApiClient) ServiceRemove(ctx context.Context, serviceID string) error {
	return nil
}

func (c testServiceApiClient) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error) {
	return types.ServiceUpdateResponse{}, nil
}

func (c testServiceApiClient) ServiceLogs(ctx context.Context, serviceID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (c testServiceApiClient) TaskLogs(ctx context.Context, taskID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (c testServiceApiClient) TaskInspectWithRaw(ctx context.Context, taskID string) (swarm.Task, []byte, error) {
	return swarm.Task{}, []byte{}, nil
}

func (c testServiceApiClient) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	return c.serviceListTasks, c.serviceListError
}

// testNodeApiClient

type testNodeApiClient struct {
	nodeListNodes []swarm.Node
	nodeListError error
}

func (c testNodeApiClient) NodeInspectWithRaw(ctx context.Context, nodeID string) (swarm.Node, []byte, error) {
	return swarm.Node{}, []byte{}, nil
}

func (c testNodeApiClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	return c.nodeListNodes, c.nodeListError
}

func (c testNodeApiClient) NodeRemove(ctx context.Context, nodeID string, options types.NodeRemoveOptions) error {
	return nil
}

func (c testNodeApiClient) NodeUpdate(ctx context.Context, nodeID string, version swarm.Version, node swarm.NodeSpec) error {
	return nil
}

func TestReaderSuccessReturnsOK(t *testing.T) {
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: []swarm.Service{},
		serviceListError:    nil,
		serviceListTasks:    nil,
	}
	n := &testNodeApiClient{
		nodeListNodes: []swarm.Node{},
		nodeListError: nil,
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	expected := http.StatusOK
	if status := w.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expected)
	}
}

func TestReaderSuccessReturnsJsonContent(t *testing.T) {
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: []swarm.Service{},
		serviceListError:    nil,
		serviceListTasks:    nil,
	}
	n := &testNodeApiClient{
		nodeListNodes: []swarm.Node{},
		nodeListError: nil,
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	expected := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("content type header does not match: got %v want %v",
			contentType, expected)
	}
}

func TestReaderSuccessReturnsCorrectBodyWithZeroFunctions(t *testing.T) {
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: []swarm.Service{},
		serviceListError:    nil,
		serviceListTasks:    nil,
	}
	n := &testNodeApiClient{
		nodeListNodes: []swarm.Node{},
		nodeListError: nil,
	}

	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	expected := "[]"
	if w.Body.String() != expected {
		t.Errorf("handler returned wrong body: got %v want %v",
			w.Body.String(), expected)
	}
}

func TestReaderSuccessReturnsCorrectBodyWithOneFunction(t *testing.T) {
	replicas := uint64(5)
	labels := map[string]string{
		"function": "bar",
	}

	services := []swarm.Service{
		swarm.Service{
			Spec: swarm.ServiceSpec{
				Mode: swarm.ServiceMode{
					Replicated: &swarm.ReplicatedService{
						Replicas: &replicas,
					},
				},
				Annotations: swarm.Annotations{
					Name:   "bar",
					Labels: labels,
				},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: swarm.ContainerSpec{
						Env: []string{
							"fprocess=bar",
						},
						Image:  "foo/bar:latest",
						Labels: labels,
					},
				},
			},
		},
	}
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: services,
		serviceListError:    nil,
		serviceListTasks:    nil,
	}
	n := &testNodeApiClient{
		nodeListNodes: []swarm.Node{},
		nodeListError: nil,
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	functions := []requests.Function{
		requests.Function{
			Name:            "bar",
			Image:           "foo/bar:latest",
			InvocationCount: 0,
			Replicas:        5,
			EnvProcess:      "bar",
			Labels: &map[string]string{
				"function": "bar",
			},
		},
	}

	marshalled, _ := json.Marshal(functions)
	expected := string(marshalled)
	if w.Body.String() != expected {
		t.Errorf("handler returned wrong body: got %v want %v",
			w.Body.String(), expected)
	}
}

func TestReaderSuccessReturnsCorrectBodyWithOneFunctionVerbose(t *testing.T) {
	replicas := uint64(1)
	labels := map[string]string{
		"function": "bar",
	}

	services := []swarm.Service{
		swarm.Service{
			Spec: swarm.ServiceSpec{
				Mode: swarm.ServiceMode{
					Replicated: &swarm.ReplicatedService{
						Replicas: &replicas,
					},
				},
				Annotations: swarm.Annotations{
					Name:   "bar",
					Labels: labels,
				},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: swarm.ContainerSpec{
						Env: []string{
							"fprocess=bar",
						},
						Image:  "foo/bar:latest",
						Labels: labels,
					},
				},
			},
		},
	}
	m := metrics.MetricOptions{}
	tasks := []swarm.Task{
		swarm.Task{
			NodeID: "manager",
			Annotations: swarm.Annotations{
				Name:   "testTask",
				Labels: map[string]string{},
			},
			Spec: swarm.TaskSpec{},
			Status: swarm.TaskStatus{
				State: swarm.TaskStateRunning,
			},
		},
	}
	nodes := []swarm.Node{
		swarm.Node{
			ID: "manager",
			Spec: swarm.NodeSpec{
				Annotations: swarm.Annotations{
					Name:   "testNode",
					Labels: map[string]string{},
				},
				Role:         "manager",
				Availability: "active",
			},
		},
	}
	c := &testServiceApiClient{
		serviceListServices: services,
		serviceListError:    nil,
		serviceListTasks:    tasks,
	}
	n := &testNodeApiClient{
		nodeListNodes: nodes,
		nodeListError: nil,
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "system/functions?v=true", nil)
	if err != nil {
		t.Fatalf("Error creating get request: %v\n", err)
	}
	handler.ServeHTTP(w, r)

	functions := []requests.Function{
		requests.Function{
			Name:            "bar",
			Image:           "foo/bar:latest",
			InvocationCount: 0,
			Replicas:        1,
			ReplicaCount:    1,
			EnvProcess:      "bar",
			Labels: &map[string]string{
				"function": "bar",
			},
		},
	}
	marshalled, _ := json.Marshal(functions)
	expected := string(marshalled)
	if w.Body.String() != expected {
		t.Errorf("handler returned wrong body: got %v want %v",
			w.Body.String(), expected)
	}
}

func TestReaderErrorReturnsInternalServerError(t *testing.T) {
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: nil,
		serviceListError:    errors.New("error"),
	}
	n := &testNodeApiClient{
		nodeListNodes: nil,
		nodeListError: errors.New("error"),
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	expected := http.StatusInternalServerError
	if status := w.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expected)
	}
}

func TestReaderErrorReturnsCorrectBody(t *testing.T) {
	m := metrics.MetricOptions{}
	c := &testServiceApiClient{
		serviceListServices: nil,
		serviceListError:    errors.New("error"),
	}
	n := &testNodeApiClient{
		nodeListNodes: nil,
		nodeListError: errors.New("error"),
	}
	handler := handlers.MakeFunctionReader(m, c, n)

	w := httptest.NewRecorder()
	r := &http.Request{}
	handler.ServeHTTP(w, r)

	expected := "Error getting service list"
	if w.Body.String() != expected {
		t.Errorf("handler returned wrong body: got %v want %v",
			w.Body.String(), expected)
	}
}
