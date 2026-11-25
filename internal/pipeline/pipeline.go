package pipeline

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/internal/newrelic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/output"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/reasoner"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/sumologic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/telemetry"
	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

const (
	logLimit       = 200
	traceLimit     = 10
	postAlertGrace = time.Minute
)

// Pipeline runs one alert end to end: New Relic, Sumo Logic, Claude, then email
// (or stdout in dry-run mode).
type Pipeline struct {
	NewRelic      *newrelic.Client
	Sumo          *sumologic.Client
	LLM           *reasoner.Client
	Mailer        *output.Mailer
	ServiceField  string
	DefaultWindow time.Duration
	DryRun        bool
	Out           io.Writer
}

// Run processes one alert. Failures in New Relic or Sumo are recorded as
// warnings and the analysis proceeds with whatever data was collected.
func (p *Pipeline) Run(ctx context.Context, alert models.Alert) error {
	started := time.Now()
	defer func() { telemetry.RecordAgentDuration(ctx, time.Since(started)) }()

	w := p.window(alert)
	nrWindow := newrelic.Window{Start: w.Start, End: w.End}

	slog.Info("pipeline.start", "alert_id", alert.ID, "service", alert.Service,
		"window_start", w.Start, "window_end", w.End)

	bundle := reasoner.Bundle{Alert: alert, Start: w.Start, End: w.End}

	t := time.Now()
	metrics, err := p.NewRelic.ServiceMetrics(alert.Service, nrWindow)
	telemetry.RecordNewRelicDuration(ctx, time.Since(t), "metrics")
	if err != nil {
		bundle.Warnings = append(bundle.Warnings, "New Relic metrics unavailable: "+err.Error())
	}
	bundle.Metrics = metrics

	t = time.Now()
	traces, err := p.NewRelic.SlowTraces(alert.Service, nrWindow, traceLimit)
	telemetry.RecordNewRelicDuration(ctx, time.Since(t), "traces")
	if err != nil {
		bundle.Warnings = append(bundle.Warnings, "New Relic traces unavailable: "+err.Error())
	}
	bundle.Traces = traces

	query := fmt.Sprintf(`%s="%s" (error OR exception OR warn OR timeout OR fatal)`, p.ServiceField, alert.Service)
	t = time.Now()
	logs, err := p.Sumo.Search(query, w.Start, w.End, logLimit)
	telemetry.RecordSumoLogicDuration(ctx, time.Since(t))
	if err != nil {
		bundle.Warnings = append(bundle.Warnings, "Sumo Logic logs unavailable: "+err.Error())
	}
	sort.SliceStable(logs, func(i, j int) bool { return logs[i].Timestamps.Before(logs[j].Timestamps) })
	bundle.Logs = logs

	t = time.Now()
	analysis, err := p.LLM.Analyze(ctx, reasoner.BuildPrompt(bundle))
	telemetry.RecordLLMDuration(ctx, time.Since(t))
	if err != nil {
		return fmt.Errorf("llm analysis: %w", err)
	}

	subject, body := output.RenderReport(alert.Name, alert.Service, w.Start, w.End, analysis, bundle.Warnings)
	if p.DryRun {
		fmt.Fprintf(p.Out, "Subject: %s\n\n%s", subject, body)
	} else if err := p.Mailer.Send(subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	slog.Info("pipeline.done", "alert_id", alert.ID, "service", alert.Service,
		"duration", time.Since(started), "log_events", len(bundle.Logs), "warnings", len(bundle.Warnings))
	return nil
}

// window uses the alert's own range when it has one, otherwise the default
// look-back ending just after the alert fired.
func (p *Pipeline) window(a models.Alert) newrelic.Window {
	if !a.WindowStart.IsZero() && !a.WindowEnd.IsZero() {
		return newrelic.Window{Start: a.WindowStart, End: a.WindowEnd}
	}
	anchor := a.Timestamp
	if anchor.IsZero() {
		anchor = time.Now()
	}
	return newrelic.Window{Start: anchor.Add(-p.DefaultWindow), End: anchor.Add(postAlertGrace)}
}
