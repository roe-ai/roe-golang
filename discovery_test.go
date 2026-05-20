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

	if path != "/v1/agents/types/" {
		t.Fatalf("unexpected path: %s", path)
	}
	if len(result.EngineTypes) != 1 || result.EngineTypes[0] != "ResearchEngine" {
		t.Fatalf("unexpected engine types: %#v", result.EngineTypes)
	}
}

func TestDiscoveryListSupportedModelsWithCapability(t *testing.T) {
	var capability string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents/models/" {
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

// roe-main PR 3232 trimmed the public engine payload to six fields. This
// guards against the SDK silently regressing if a future regen reintroduces
// typed fields the backend no longer returns (e.g. form_type).
func TestDiscoveryListAgentEngineTypesParsesPublicEnginePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"engine_types": []string{"ResearchEngine"},
			"total_count":  1,
			"engines": []map[string]any{
				{
					"class_id":       "ResearchEngine",
					"display_name":   "Research Engine",
					"description":    "Researches things.",
					"summary":        "Research workflow.",
					"input_schema":   map[string]any{"type": "object", "properties": map[string]any{}},
					"default_values": map[string]any{},
				},
			},
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

	if len(result.Engines) != 1 {
		t.Fatalf("expected 1 engine, got %d", len(result.Engines))
	}
	engine := result.Engines[0]
	if engine["class_id"] != "ResearchEngine" {
		t.Fatalf("unexpected class_id: %v", engine["class_id"])
	}
	if engine["display_name"] != "Research Engine" {
		t.Fatalf("unexpected display_name: %v", engine["display_name"])
	}
	if _, ok := engine["input_schema"].(map[string]any); !ok {
		t.Fatalf("input_schema is not a map: %#v", engine["input_schema"])
	}
	if _, ok := engine["default_values"].(map[string]any); !ok {
		t.Fatalf("default_values is not a map: %#v", engine["default_values"])
	}
}

func TestDiscoveryListSupportedModelsParsesPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Greptile P2: guard that an empty capability is not forwarded as
		// `?capability=`, which a future refactor could accidentally introduce.
		if r.URL.Query().Has("capability") {
			t.Errorf("expected no capability query param, got %q", r.URL.Query().Get("capability"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]any{
				{
					"id":                        "gpt-5",
					"providers":                 []string{"openai"},
					"capabilities":              []string{"text"},
					"context_window":            200000,
					"max_output_tokens":         8192,
					"supports_system_message":   true,
					"supports_temperature":      true,
					"supports_reasoning_effort": false,
					"supports_json_output":      true,
					"supports_json_schema":      true,
				},
			},
			"total_count":  1,
			"tenant_scope": "all_tenants",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", "test-org", server.URL, 0, 0)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	result, err := client.Discovery.ListSupportedModels("")
	if err != nil {
		t.Fatalf("ListSupportedModels returned error: %v", err)
	}

	if result.TotalCount != 1 {
		t.Fatalf("unexpected total_count: %d", result.TotalCount)
	}
	if len(result.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Models))
	}
	if result.Models[0].Id != "gpt-5" {
		t.Fatalf("unexpected model id: %s", result.Models[0].Id)
	}
}
