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

func TestHTTPClientNegativeMaxRetriesStillSendsRequest(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := Config{APIKey: "k", OrganizationID: "org", BaseURL: server.URL, Timeout: time.Second, MaxRetries: -1}
	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	var out map[string]bool
	if err := client.get("/ok", nil, &out); err != nil || !out["ok"] {
		t.Fatalf("get = %v, %v", out, err)
	}
}

func TestAgentRunRequestsAreNotRetried(t *testing.T) {
	requests := 0
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if _, present := r.Header[skipRetryHeader]; present {
			t.Errorf("%s header must not be sent", skipRetryHeader)
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{APIKey: "k", OrganizationID: "org", BaseURL: server.URL, Timeout: time.Second, MaxRetries: 3})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	runs := map[string]func() error{
		"Run": func() error {
			_, err := client.Agents.Run("agent-id", 0, map[string]any{"text": "hi"}, nil)
			return err
		},
		"RunMany": func() error {
			_, err := client.Agents.RunMany("agent-id", []map[string]any{{"text": "hi"}}, 0, nil)
			return err
		},
	}
	for name, run := range runs {
		requests = 0
		if err := run(); err == nil {
			t.Fatalf("%s: expected error", name)
		}
		if requests != 1 {
			t.Fatalf("%s: expected 1 request, got %d", name, requests)
		}
	}
}
