package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/config"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/newrelic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/output"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/pipeline"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/reasoner"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/sumologic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/telemetry"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/webhook"
	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

const pipelineTimeout = 10 * time.Minute

func main() {
	dryRun := flag.Bool("dry-run", false, "print the report to stdout instead of sending email")
	flag.Parse()
	if *dryRun {
		_ = os.Setenv("DRY_RUN", "true")
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config.invalid", "error", err)
		os.Exit(1)
	}

	if cfg.OTelEnabled {
		shutdown := telemetry.Init(cfg)
		defer shutdown()
	}

	pl := &pipeline.Pipeline{
		NewRelic:      newrelic.NewClient(),
		Sumo:          sumologic.NewClient(cfg.SumoEndpoint),
		LLM:           reasoner.NewClient(cfg.AnthropicAPIKey, cfg.AnthropicModel),
		Mailer:        output.NewMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.EmailFrom, cfg.EmailTo),
		ServiceField:  cfg.SumoServiceField,
		DefaultWindow: cfg.QueryWindow,
		DryRun:        cfg.DryRun,
		Out:           os.Stdout,
	}

	handler := webhook.NewHandler(cfg.WebhookSecret, func(a models.Alert) {
		ctx, cancel := context.WithTimeout(context.Background(), pipelineTimeout)
		defer cancel()
		if err := pl.Run(ctx, a); err != nil {
			slog.Error("pipeline.failed", "alert_id", a.ID, "service", a.Service, "error", err)
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", handler.ServeHTTP)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.WebhookPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("agent.listening", "port", cfg.WebhookPort, "dry_run", cfg.DryRun, "service", cfg.ServiceName)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server.failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	slog.Info("agent.stopped")
}
