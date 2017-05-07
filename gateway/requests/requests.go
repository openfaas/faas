// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package requests

// CreateFunctionRequest create a function in the swarm.
type CreateFunctionRequest struct {
	Service    string `json:"service"`
	Image      string `json:"image"`
	Network    string `json:"network"`

	// EnvProcess corresponds to the fprocess variable for your container watchdog.
	EnvProcess string `json:"envProcess"`

	// EnvVars provides overrides for functions.
	EnvVars map[string]string `json:"envVars"`
}

type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

type AlexaSessionApplication struct {
	ApplicationId string `json:"applicationId"`
}

type AlexaSession struct {
	SessionId   string                  `json:"sessionId"`
	Application AlexaSessionApplication `json:"application"`
}

type AlexaIntent struct {
	Name string `json:"name"`
}

type AlexaRequest struct {
	Intent AlexaIntent `json:"intent"`
}

// AlexaRequestBody top-level request produced by Alexa SDK
type AlexaRequestBody struct {
	Session AlexaSession `json:"session"`
	Request AlexaRequest `json:"request"`
}

type PrometheusInnerAlertLabel struct {
	AlertName    string `json:"alertname"`
	FunctionName string `json:"function_name"`
}

type PrometheusInnerAlert struct {
	Status string                    `json:"status"`
	Labels PrometheusInnerAlertLabel `json:"labels"`
}

// PrometheusAlert as produced by AlertManager
type PrometheusAlert struct {
	Status   string                 `json:"status"`
	Receiver string                 `json:"receiver"`
	Alerts   []PrometheusInnerAlert `json:"alerts"`
}

// Function exported for system/functions endpoint
type Function struct {
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	InvocationCount float64 `json:"invocationCount"`
	Replicas        uint64  `json:"replicas"`
}
