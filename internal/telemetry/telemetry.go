package telemetry
import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"

    "github.com/fadikaba81/ObservabilityAgentAI/config"
)

func Init(cfg *config.Config) func() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	res, err := resource.New(context.Background(), 
	resource.WithAttributes(
		attribute.String("serivce.name", cfg.ServiceName),
		attribute/String("deployment.environment", cfg.Environment),
	),
}
