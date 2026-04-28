package roe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestRawClientRetriesOn503 verifies the generated raw client gets the SDK's
// retry policy via the retryDoer adapter — a 503 followed by a 200 should
// succeed transparently.
func TestRawClientRetriesOn503(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:               "k",
		OrganizationID:       "test-org",
		BaseURL:              server.URL,
		Timeout:              2 * time.Second,
		MaxRetries:           2,
		RetryInitialInterval: time.Millisecond,
		RetryMaxInterval:     time.Millisecond,
		RetryMultiplier:      1,
	})
	if err != nil {
		t.Fatalf("NewClientWithConfig: %v", err)
	}
	defer client.Close()

	raw, err := client.Raw()
	if err != nil {
		t.Fatalf("Raw: %v", err)
	}

	resp, err := raw.V1UsersCurrentUserRetrieveWithResponse(context.Background())
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200 after retry, got %d", resp.StatusCode())
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("expected 2 attempts (1 fail + 1 success), got %d", got)
	}
}

// TestRawClientExhaustsRetries verifies that persistent 503s exhaust the
// retry budget and surface the final response so the caller can observe the
// status code.
func TestRawClientExhaustsRetries(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:               "k",
		OrganizationID:       "test-org",
		BaseURL:              server.URL,
		Timeout:              2 * time.Second,
		MaxRetries:           2,
		RetryInitialInterval: time.Millisecond,
		RetryMaxInterval:     time.Millisecond,
		RetryMultiplier:      1,
	})
	if err != nil {
		t.Fatalf("NewClientWithConfig: %v", err)
	}
	defer client.Close()

	raw, err := client.Raw()
	if err != nil {
		t.Fatalf("Raw: %v", err)
	}

	resp, err := raw.V1UsersCurrentUserRetrieveWithResponse(context.Background())
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if resp.StatusCode() != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after exhausting retries, got %d", resp.StatusCode())
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("expected 3 attempts (MaxRetries=2 + 1), got %d", got)
	}
}
