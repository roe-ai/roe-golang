package roe

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAgentsAPIListWithContextSendsQueryAndSetsAgentAPI(t *testing.T) {
	var calls int32
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("organization_id"); got != "org" {
			t.Fatalf("expected organization_id=org, got %q", got)
		}
		if got := q.Get("page"); got != "1" {
			t.Fatalf("expected page=1, got %q", got)
		}
		if got := q.Get("page_size"); got != "10" {
			t.Fatalf("expected page_size=10, got %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Fatalf("expected auth header set, got %q", got)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Fatalf("expected user-agent header set")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{"id":"a1","name":"Agent","disable_cache":false,"cache_failed_jobs":false,"organization_id":"org","engine_class_id":"engine","job_count":0,"engine_name":"Engine"}]}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: "org",
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.Agents.ListWithContext(ctx, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].agentsAPI == nil {
		t.Fatalf("expected agentsAPI to be set on result")
	}
	if resp.Results[0].agentsAPI != client.Agents {
		t.Fatalf("expected result agentsAPI to match client agents API")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", atomic.LoadInt32(&calls))
	}
}

func TestAgentsAPIListWithContextCancelledDoesNotSendRequest(t *testing.T) {
	var calls int32
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: "org",
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Agents.ListWithContext(ctx, 1, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expected no server calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestAgentsAPIRunManyWithContextStopsAfterCancel(t *testing.T) {
	var calls int32
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.URL.Path != "/v1/agents/run/agent-id/async/many/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["job-1"]`))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var hookCalls int32
	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: "org",
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
		AfterResponse: []ResponseHook{
			func(resp *http.Response, _ []byte) {
				if resp == nil || resp.Request == nil {
					return
				}
				if !strings.Contains(resp.Request.URL.Path, "/async/many/") {
					return
				}
				if atomic.AddInt32(&hookCalls, 1) == 1 {
					cancel()
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	inputs := make([]map[string]any, 1001)
	for i := range inputs {
		inputs[i] = map[string]any{"text": "hello"}
	}

	_, err = client.Agents.RunManyWithContext(ctx, "agent-id", inputs, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 call, got %d", atomic.LoadInt32(&calls))
	}
}
