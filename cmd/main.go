package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fadikaba81/ObservabilityAgentAI/config"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/telemetry"

	"github.com/fadikaba81/ObservabilityAgentAI/internal/newrelic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/sumologic"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/webhook"
)

func main() {
	cfg := config.Load()

	shutdown := telemetry.Init(cfg)
	defer shutdown()

	err := telemetry.InitMetrics()

	if err != nil {
		slog.Error("failed to init metrics", "error", err)
		os.Exit(1)
	}

	sumoClient := sumologic.NewClient(cfg.SumoEndpoint)
	newrelicClient := newrelic.NewClient()

	handler := webhook.NewHandler(sumoClient, newrelicClient)

	http.HandleFunc("/webhook", handler.HandleAlert)

	slog.Info("ObsAIAgent Starting",
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
	)

	go http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebhookPort), nil)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("ObsAIAgent shutting down")

}
