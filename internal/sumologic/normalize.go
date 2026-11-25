package sumologic

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

// NormalizeMessages converts Sumo Logic messages into events, sorted oldest first.
func NormalizeMessages(messages []SearchJobMessage) ([]models.Event, error) {
	events := make([]models.Event, 0, len(messages))

	for _, msg := range messages {
		raw := msg.Map["_raw"]
		events = append(events, models.Event{
			ID:             msg.Map["_messageid"],
			Timestamps:     messageTime(msg.Map["_messagetime"]),
			Service:        msg.Map["service"],
			Message:        raw,
			Severity:       severityOf(raw),
			Source:         models.SourceSumoLogic,
			RawLog:         raw,
			SourceCategory: msg.Map["_sourceCategory"],
			TraceID:        firstNonEmpty(msg.Map["trace_id"], msg.Map["traceId"], msg.Map["_traceId"]),
		})
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamps.Before(events[j].Timestamps)
	})
	return events, nil
}

// messageTime reads Sumo's _messagetime (epoch milliseconds). It returns the
// zero time when the field is missing, rather than the time of processing.
func messageTime(v string) time.Time {
	ms, err := strconv.ParseInt(v, 10, 64)
	if err != nil || ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

func severityOf(raw string) models.Severity {
	l := strings.ToLower(raw)
	switch {
	case strings.Contains(l, "fatal") || strings.Contains(l, "critical"):
		return models.SeverityCritical
	case strings.Contains(l, "error") || strings.Contains(l, "exception"):
		return models.SeverityError
	case strings.Contains(l, "warn"):
		return models.SeverityWarning
	default:
		return models.SeverityInfo
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}