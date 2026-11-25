package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Histograms are nil until InitMetrics runs (i.e. when OTEL_ENABLED=true).
// The Record functions are no-ops in that case.
var (
	sumoQueryDuration     metric.Float64Histogram
	newrelicQueryDuration metric.Float64Histogram
	llmRequestDuration    metric.Float64Histogram
	agentRunDuration      metric.Float64Histogram
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
	newrelicQueryDuration, err = m.Float64Histogram(
		"newrelic.query.duration",
		metric.WithDescription("Time taken to run an NRQL query through NerdGraph"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	llmRequestDuration, err = m.Float64Histogram(
		"llm.query.duration",
		metric.WithDescription("Time taken to complete an LLM request"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	agentRunDuration, err = m.Float64Histogram(
		"agent.run.duration",
		metric.WithDescription("Time taken for one full alert run"),
		metric.WithUnit("s"),
	)
	return err
}

// RecordSumoLogicDuration records how long a Sumo Logic search job took.
func RecordSumoLogicDuration(ctx context.Context, duration time.Duration) {
	if sumoQueryDuration == nil {
		return
	}
	sumoQueryDuration.Record(ctx, duration.Seconds())
}

// RecordNewRelicDuration records one NRQL query. queryName separates the
// queries, e.g. "metrics" or "traces".
func RecordNewRelicDuration(ctx context.Context, duration time.Duration, queryName string) {
	if newrelicQueryDuration == nil {
		return
	}
	newrelicQueryDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("query", queryName)))
}

// RecordLLMDuration records how long the LLM request took.
func RecordLLMDuration(ctx context.Context, duration time.Duration) {
	if llmRequestDuration == nil {
		return
	}
	llmRequestDuration.Record(ctx, duration.Seconds())
}

// RecordAgentDuration records the end-to-end time for one alert.
func RecordAgentDuration(ctx context.Context, duration time.Duration) {
	if agentRunDuration == nil {
		return
	}
	agentRunDuration.Record(ctx, duration.Seconds())
}