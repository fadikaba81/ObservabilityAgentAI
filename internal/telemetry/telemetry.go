package telemetry

import (
	"context"
	"log/slog"
	"os"

	"github.com/fadikaba81/ObservabilityAgentAI/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Init(cfg *config.Config) func() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("serivce.name", cfg.ServiceName),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)

	if err != nil {
		slog.Error("Failed to create otel resource", "error", err)
	}

	shutdownTrace := initTracer(cfg, res)
	shutdownMeter := initMeter(cfg, res)

	slog.Info("telemetry initialised",
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
		"endpoint", cfg.OTelExporter,
	)

	return func() {
		shutdownTrace()
		shutdownMeter()
	}

}

func initTracer(cfg *config.Config, res *resource.Resource) func() {
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTelExporter),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		slog.Error("failed to init tracer exporter", "error", err)
		return func() {}
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer provider", "error", err)
		}
	}

}
func initMeter(cfg *config.Config, res *resource.Resource) func() {
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(cfg.OTelEndpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		slog.Error("failed to init metrics exporter", "error", err)
		return func() {}
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	return func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown meter provider", "error", err)
		}
	}

}
