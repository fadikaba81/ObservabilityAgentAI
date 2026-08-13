package sumologic

import (
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

func NormalizeMessages(message []SearchJobMessage) ([]models.Event, error) {
	events := make([]models.Event, 0, len(message))

	for _, msg := range message {
		event := models.Event{
			Timestamps:     time.Now(),
			Service:        msg.Map["service"],
			Message:        msg.Map["_raw"],
			Severity:       models.SeverityInfo,
			Source:         models.SourceSumoLogic,
			RawLog:         msg.Map["_raw"],
			SourceCategory: msg.Map["_sourceCategory"],
			TraceID:        msg.Map["_traceId"],
		}
		events = append(events, event)
	}

	return events, nil

}
