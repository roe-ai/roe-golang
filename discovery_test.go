package roe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoveryListAgentEngineTypes(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"engine_types": []string{"ResearchEngine"},
			"total_count":  1,
			"engines":      []any{},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", "test-org", server.URL, 0, 0)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	result, err := client.Discovery.ListAgentEngineTypes()
	if err != nil {
		t.Fatalf("ListAgentEngineTypes returned error: %v", err)
	}

	if path != "/v1/discovery/agent-engine-types/" {
		t.Fatalf("unexpected path: %s", path)
	}
	if len(result.EngineTypes) != 1 || result.EngineTypes[0] != "ResearchEngine" {
		t.Fatalf("unexpected engine types: %#v", result.EngineTypes)
	}
}

func TestDiscoveryListSupportedModelsWithCapability(t *testing.T) {
	var capability string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/discovery/supported-models/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		capability = r.URL.Query().Get("capability")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models":       []any{},
			"total_count":  0,
			"tenant_scope": "all_tenants",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", "test-org", server.URL, 0, 0)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	result, err := client.Discovery.ListSupportedModels("image")
	if err != nil {
		t.Fatalf("ListSupportedModels returned error: %v", err)
	}

	if capability != "image" {
		t.Fatalf("unexpected capability query: %s", capability)
	}
	if result.TenantScope != "all_tenants" {
		t.Fatalf("unexpected tenant scope: %s", result.TenantScope)
	}
}
