package config

import "os"

type Config struct {
	ServiceName  string
	Environment  string
	OTelExporter string
	OTelEndpoint string
}

func Load() *Config {
	return &Config{
		ServiceName:  getEnv("SERVICE_NAME", "ObsAIAgent"),
		Environment:  getEnv("ENVIRONMENT", "dev"),
		OTelExporter: getEnv("OTEL_EXPORTER", "http"),
		OTelEndpoint: getEnv("OTEL_ENDPOINT", "http://localhost:4318"),
	}
}
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
