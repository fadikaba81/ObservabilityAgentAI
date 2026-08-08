package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	sumoQueryDuration  metric.Float64Histogram
	llmRequestDuration metric.Float64Histogram
	agentRunDuration   metric.Float64Histogram
)

func InitMetrics() error {

	var err error

	m := otel.Meter("observabilityagentai")

	sumoQueryDuration, err = m.Float64Histogram(
		"sumo.query.duration",
		metric.WithDescription("Time taken to complete a Sumo Logic search job"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	llmRequestDuration, err = m.Float64Histogram(
		"llm.query.duration",
		metric.WithDescription("Time taken to complete LLM search"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	agentRunDuration, err = m.Float64Histogram(
		"agent.run.duration",
		metric.WithDescription("Time taken to complete agent search"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	return nil
}

func RecordSumoLogicDuration(cxt context.Context, duration time.Duration) {
	sumoQueryDuration.Record(cxt, duration.Seconds())
}
func RecordLLMDuration(cxt context.Context, duration time.Duration) {
	llmRequestDuration.Record(cxt, duration.Seconds())
}
func RecordAgentDuration(cxt context.Context, duration time.Duration) {
	agentRunDuration.Record(cxt, duration.Seconds())
}
