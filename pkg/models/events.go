package models

import "time"

type Source string

const (
	SourceSumoLogic Source = "sumologic"
	SourceNewRelic  Source = "newrelic"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type Event struct {
	// Common fields
	ID         string    `json:"id"`
	Timestamps time.Time `json:"timestamps"`
	Service    string    `json:"service"`
	Severity   Severity  `json:"severity"`
	Message    string    `json:"message"`
	Source     Source    `json:"source"`

	//Tracing (NewRelic)
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`

	//Metric (NewRelic)
	MetricsName  string  `json:"metrics_name,omitempty"`
	MetricsValue float64 `json:"metrics_value,omitempty"`

	// Logs (SumoLogic)
	RawLog         string `json:"raw_log,omitempty"`
	SourceCategory string `json:"source_category,omitempty"`

	//Extra k/v for anything that doesn't fit above
	Attributes map[string]string `json:"attributes,omitempty"`
}
