package config

import (
	"testing"
)

func TestLoadWithDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	cfg, err := Load()
	if err == nil {
		t.Error("Should fail without DATABASE_URL")
	}
	_ = cfg
}

func TestLoadWithEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/fms")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}

	if cfg.App.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.App.Port)
	}
	if cfg.App.LogLevel != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.App.LogLevel)
	}
}

func TestGetKafkaBrokers(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/fms")
	t.Setenv("KAFKA_BROKERS", "localhost:9092,localhost:9093")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}

	brokers := cfg.GetKafkaBrokers()
	if len(brokers) != 2 {
		t.Errorf("Expected 2 brokers, got %d", len(brokers))
	}
}