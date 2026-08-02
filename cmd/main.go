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

	slog.Info("ObsAIAgent Starting",
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("ObsAIAgent shutting down")

}
