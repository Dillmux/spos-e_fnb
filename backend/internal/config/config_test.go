package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg := LoadConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Port == "" {
		t.Error("expected non-empty Port")
	}
}

func TestLoadDotEnv(t *testing.T) {
	// Create temporary env file
	tmpFile, err := os.CreateTemp("", ".env.*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "TEST_ENV_VAR=hello_world\n# Comment line\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	loadDotEnv(tmpFile.Name())

	if val := os.Getenv("TEST_ENV_VAR"); val != "hello_world" {
		t.Errorf("expected TEST_ENV_VAR to be 'hello_world', got '%s'", val)
	}
}
