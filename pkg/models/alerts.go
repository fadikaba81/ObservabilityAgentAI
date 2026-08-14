package models

import "time"

type Alert struct {
	ID           string    `json:"id"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	Description  string    `json:"description"`
	ErrorMessage string    `json:"error_message"`
	Service      string    `json:"service"`
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
}
