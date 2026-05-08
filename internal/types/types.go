package types

type HealthResponse struct {
	Status string `json:"status"`
}

type DependencyStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	Error string `json:"error,omitempty"`
}

type ReadyResponse struct {
	Status       string             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

type PingResponse struct {
	Message   string `json:"message"`
	TraceID   string `json:"traceId"`
	RequestID string `json:"requestId"`
}
