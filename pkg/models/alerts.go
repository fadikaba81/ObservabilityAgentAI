package models

import "time"

// Alert is the payload posted by a Sumo Logic webhook connection.
// The payload template should send timestamps as RFC 3339 (ISO 8601).
type Alert struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	Description  string    `json:"description"`
	ErrorMessage string    `json:"error_message"`
	Service      string    `json:"service"`
	Timestamp    time.Time `json:"timestamp"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Source       string    `json:"source"`
}
