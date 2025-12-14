package roe

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigEnvParsing(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ROE_API_KEY":           "test-key",
		"ROE_ORGANIZATION_ID":   "org-123",
		"ROE_TIMEOUT":           "90s",
		"ROE_MAX_RETRIES":       "5",
		"ROE_DEBUG":             "true",
		"ROE_PROXY":             "http://localhost:8080",
		"ROE_EXTRA_HEADERS":     "X-Test=one;X-Another:two",
		"ROE_REQUEST_ID":        "req-abc",
		"ROE_REQUEST_ID_HEADER": "X-Custom-Request-ID",
		"ROE_RETRY_INITIAL_MS":  "50",
		"ROE_RETRY_MAX_MS":      "150",
		"ROE_RETRY_MULTIPLIER":  "1.5",
		"ROE_RETRY_JITTER":      "0.1",
	})
	defer restore()

	cfg, err := LoadConfig("", "", "", 0, 0)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Timeout != 90*time.Second {
		t.Fatalf("expected timeout 90s, got %s", cfg.Timeout)
	}
	if cfg.MaxRetries != 5 {
		t.Fatalf("expected max retries 5, got %d", cfg.MaxRetries)
	}
	if !cfg.Debug {
		t.Fatalf("expected debug to be true")
	}
	if cfg.ProxyURL == nil || cfg.ProxyURL.String() != "http://localhost:8080" {
		t.Fatalf("expected proxy url set, got %v", cfg.ProxyURL)
	}
	if cfg.ExtraHeaders.Get("X-Test") != "one" || cfg.ExtraHeaders.Get("X-Another") != "two" {
		t.Fatalf("unexpected extra headers: %v", cfg.ExtraHeaders)
	}
	if cfg.DefaultRequestID != "req-abc" {
		t.Fatalf("expected request id req-abc, got %s", cfg.DefaultRequestID)
	}
	if cfg.RequestIDHeader != "X-Custom-Request-ID" {
		t.Fatalf("expected custom request id header, got %s", cfg.RequestIDHeader)
	}
	if cfg.RetryInitialInterval != 50*time.Millisecond || cfg.RetryMaxInterval != 150*time.Millisecond {
		t.Fatalf("unexpected retry intervals: %s %s", cfg.RetryInitialInterval, cfg.RetryMaxInterval)
	}
	if cfg.RetryMultiplier != 1.5 {
		t.Fatalf("expected retry multiplier 1.5, got %f", cfg.RetryMultiplier)
	}
	if cfg.RetryJitter != 0.1 {
		t.Fatalf("expected retry jitter 0.1, got %f", cfg.RetryJitter)
	}
}

func TestLoadConfigMaxRetriesZero(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ROE_API_KEY":         "test-key",
		"ROE_ORGANIZATION_ID": "org-123",
		"ROE_MAX_RETRIES":     "0",
	})
	defer restore()

	cfg, err := LoadConfig("", "", "", 0, 0)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.MaxRetries != 0 {
		t.Fatalf("expected max retries 0, got %d", cfg.MaxRetries)
	}
}

func TestLoadConfigInvalidIntEnvErrors(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ROE_API_KEY":         "test-key",
		"ROE_ORGANIZATION_ID": "org-123",
		"ROE_MAX_RETRIES":     "nope",
	})
	defer restore()

	_, err := LoadConfig("", "", "", 0, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadConfigValidation(t *testing.T) {
	restore := setEnvVars(map[string]string{
		"ROE_API_KEY":         "",
		"ROE_ORGANIZATION_ID": "org-123",
	})
	defer restore()

	if _, err := LoadConfig("", "", "", 0, 0); err == nil {
		t.Fatalf("expected error for missing api key")
	}
}

func setEnvVars(values map[string]string) func() {
	originals := map[string]string{}
	for k, v := range values {
		originals[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}
	return func() {
		for k, v := range originals {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}
}
