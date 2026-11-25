package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName      string
	Environment      string
	WebhookPort      int
	WebhookSecret    string
	QueryWindow      time.Duration
	SumoEndpoint     string
	SumoServiceField string
	AnthropicAPIKey  string
	AnthropicModel   string
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	EmailFrom        string
	EmailTo          string
	DryRun           bool
	OTelEnabled      bool
	OTelEndpoint     string
}

// Load reads configuration from environment variables and fails fast if
// anything required is missing.
func Load() (*Config, error) {
	window, err := time.ParseDuration(getEnv("QUERY_WINDOW", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid QUERY_WINDOW: %w", err)
	}
	webhookPort, err := strconv.Atoi(getEnv("WEBHOOK_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid WEBHOOK_PORT: %w", err)
	}
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	smtpUser := getEnv("SMTP_USERNAME", "")
	cfg := &Config{
		ServiceName:      getEnv("SERVICE_NAME", "ObsAIAgent"),
		Environment:      getEnv("ENVIRONMENT", "dev"),
		WebhookPort:      webhookPort,
		WebhookSecret:    os.Getenv("WEBHOOK_SECRET"),
		QueryWindow:      window,
		SumoEndpoint:     getEnv("SUMOLOGIC_ENDPOINT", "https://api.sumologic.com/api/v1"),
		SumoServiceField: getEnv("SUMO_SERVICE_FIELD", "service"),
		AnthropicAPIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:   getEnv("ANTHROPIC_MODEL", "claude-sonnet-5-5"),
		SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:         smtpPort,
		SMTPUsername:     smtpUser,
		SMTPPassword:     os.Getenv("SMTP_PASSWORD"),
		EmailFrom:        getEnv("EMAIL_FROM", smtpUser),
		EmailTo:          os.Getenv("EMAIL_TO"),
		DryRun:           getBool("DRY_RUN"),
		OTelEnabled:      getBool("OTEL_ENABLED"),
		OTelEndpoint:     getEnv("OTEL_ENDPOINT", "http://localhost:4318"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var missing []string

	for _, name := range []string{
		"WEBHOOK_SECRET",
		"ANTHROPIC_API_KEY",
		"SUMOLOGIC_ACCESSID_SANDBOX",
		"SUMOLOGIC_ACCESSKEY_SANDBOX",
		"NEWRELIC_APIKEY_SANDBOX",
		"NEWRELIC_ACCOUNTID_SANDBOX",
	} {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}

	if !c.DryRun {
		if c.EmailTo == "" {
			missing = append(missing, "EMAIL_TO")
		}
		if c.EmailFrom == "" {
			missing = append(missing, "EMAIL_FROM")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getBool(key string) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	return err == nil && v
}
