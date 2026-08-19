package config

import "os"

type Config struct {
	ServiceName  string
	Environment  string
	OTelExporter string
	OTelEndpoint string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	EmailFrom    string
	EmailTo      string
	WebhookPort  int
	QueryWindow  string
	SumoEndpoint string
}

func Load() *Config {
	return &Config{
		ServiceName:  getEnv("SERVICE_NAME", "ObsAIAgent"),
		Environment:  getEnv("ENVIRONMENT", "dev"),
		OTelExporter: getEnv("OTEL_EXPORTER", "http"),
		OTelEndpoint: getEnv("OTEL_ENDPOINT", "http://localhost:4318"),
		SumoEndpoint: getEnv("SUMOLOGIC_ENDPOINT", "https://api.sumologic.com/api/v1"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     587,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		EmailFrom:    getEnv("EMAIL_FROM", getEnv("SMTP_USERNAME", "")),
		EmailTo:      getEnv("EMAIL_TO", getEnv("SMTP_USERNAME", "")),
		WebhookPort:  8080,
		QueryWindow:  getEnv("QUERY_WINDOW", "5m"),
	}
}
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
