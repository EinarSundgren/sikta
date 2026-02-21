package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                         string
	DatabaseURL                  string
	AllowedOrigins               []string
	AnthropicAPIKey              string
	AnthropicAPIURL              string
	AnthropicModelExtraction     string
	AnthropicModelClassification string
	AnthropicModelChronology     string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	allowedOrigins := strings.Split(
		getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		",",
	)

	return &Config{
		Port:                        getEnv("PORT", "8080"),
		DatabaseURL:                 databaseURL,
		AllowedOrigins:               allowedOrigins,
		AnthropicAPIKey:              getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicAPIURL:              getEnv("ANTHROPIC_API_URL", "https://api.anthropic.com"),
		AnthropicModelExtraction:     getEnv("ANTHROPIC_MODEL_EXTRACTION", "claude-sonnet-4-20250514"),
		AnthropicModelClassification: getEnv("ANTHROPIC_MODEL_CLASSIFICATION", "claude-haiku-4-20250514"),
		AnthropicModelChronology:     getEnv("ANTHROPIC_MODEL_CHRONOLOGY", "claude-sonnet-4-20250514"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
