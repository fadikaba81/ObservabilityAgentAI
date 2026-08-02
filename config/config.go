package config
package config 

import "os"

type Config struct{
	ServiceName string
	Environment string
	OTelExporter string
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "ObsAIAgent"),
		Environment: getEnv("ENVIRONMENT", "dev"),
		OTelExporter: getEnv("OTEL_EXPORTER", "stdout"),
	}
}