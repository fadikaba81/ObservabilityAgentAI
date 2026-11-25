package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

var serviceNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

const dedupeTTL = time.Hour

// Handler receives Sumo Logic alert webhooks.
type Handler struct {
	secret string
	submit func(models.Alert)

	mu   sync.Mutex
	seen map[string]time.Time
}

// NewHandler returns a handler that requires secret in the X-Webhook-Secret header
// and calls submit in a goroutine for each new alert.
func NewHandler(secret string, submit func(models.Alert)) *Handler {
	return &Handler{
		secret: secret,
		submit: submit,
		seen:   make(map[string]time.Time),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.authorized(r) {
		slog.Warn("webhook.unauthorized", "remote", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var alert models.Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if !serviceNamePattern.MatchString(alert.Service) {
		http.Error(w, "service must match [A-Za-z0-9._-]", http.StatusBadRequest)
		return
	}

	key := alert.ID
	if key == "" {
		key = alert.Service + "|" + alert.Timestamp.UTC().Format(time.RFC3339)
	}
	if h.isDuplicate(key) {
		slog.Info("webhook.duplicate", "alert_id", alert.ID, "service", alert.Service)
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("webhook.accepted", "alert_id", alert.ID, "service", alert.Service, "severity", alert.Severity)
	go h.submit(alert)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) authorized(r *http.Request) bool {
	if h.secret == "" {
		return false
	}
	got := []byte(r.Header.Get("X-Webhook-Secret"))
	return subtle.ConstantTimeCompare(got, []byte(h.secret)) == 1
}

func (h *Handler) isDuplicate(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for k, t := range h.seen {
		if now.Sub(t) > dedupeTTL {
			delete(h.seen, k)
		}
	}
	if _, ok := h.seen[key]; ok {
		return true
	}
	h.seen[key] = now
	return false
}
