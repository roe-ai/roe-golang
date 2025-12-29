package roe

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHTTPClientRetriesAndRequestID(t *testing.T) {
	attempts := 0
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.Header.Get("X-Request-ID") == "" {
			t.Errorf("expected request id header to be set")
		}
		w.Header().Set("X-Request-ID", "resp-id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           2,
		RetryInitialInterval: 10 * time.Millisecond,
		RetryMaxInterval:     10 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
		AutoRequestID:        true,
		RequestIDHeader:      "X-Request-ID",
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	var out map[string]bool
	if err := client.get("/ok", nil, &out); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestAPIErrorIncludesRequestID(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "abc-123")
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"detail":"rate limited"}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 5 * time.Millisecond,
		RetryMaxInterval:     5 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
		RequestIDHeader:      "X-Request-ID",
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	err := client.get("/error", nil, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	rateErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("expected rate limit error, got %T", err)
	}
	if rateErr.RequestID != "abc-123" {
		t.Fatalf("expected request id propagated, got %s", rateErr.RequestID)
	}
	if rateErr.RetryAfter == nil || rateErr.RetryAfter.Seconds() < 0.9 {
		t.Fatalf("expected retry-after to be parsed, got %v", rateErr.RetryAfter)
	}
	if rateErr.Message != "rate limited" {
		t.Fatalf("unexpected message: %s", rateErr.Message)
	}
}

func TestHTTPClientRetrySleepHonorsContextCancellation(t *testing.T) {
	firstResponse := make(chan struct{}, 1)
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		firstResponse <- struct{}{}
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           1,
		RetryInitialInterval: 500 * time.Millisecond,
		RetryMaxInterval:     500 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-firstResponse
		cancel()
	}()

	start := time.Now()
	err := client.getWithContext(ctx, "/error", nil, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("expected cancellation to short-circuit retry sleep, took %s", elapsed)
	}
}
