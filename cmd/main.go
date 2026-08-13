package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fadikaba81/ObservabilityAgentAI/config"
	"github.com/fadikaba81/ObservabilityAgentAI/internal/telemetry"
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

	slog.Info("ObsAIAgent Starting",
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("ObsAIAgent shutting down")

}
