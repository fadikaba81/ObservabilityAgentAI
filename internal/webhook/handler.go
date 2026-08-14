package webhook

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/fadikaba81/ObservabilityAgentAI/internal/newrelic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/sumologic"
	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

type Handler struct {
	sumoClient     *sumologic.Client
	newrelicClient *newrelic.Client
}

func NewHandler(sumoClient *sumologic.Client, newrelicClient *newrelic.Client) *Handler {
	return &Handler{
		sumoClient:     sumoClient,
		newrelicClient: newrelicClient,
	}
}

func (h *Handler) HandleAlert(w http.ResponseWriter, r *http.Request) {
	var alert models.Alert

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		slog.Error("Failed to decode alert", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("alert.received",
		"service", alert.Service,
		"severity", alert.Severity,
		"message", alert.Message,
	)

	w.WriteHeader(http.StatusOK)
}
