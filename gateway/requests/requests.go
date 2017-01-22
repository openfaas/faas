package requests

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

type PrometheusAlert struct {
	Status   string                 `json:"status"`
	Receiver string                 `json:"receiver"`
	Alerts   []PrometheusInnerAlert `json:"alerts"`
}
