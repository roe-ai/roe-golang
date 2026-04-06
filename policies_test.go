package roe

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPoliciesAPIListSendsQuery(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/policies/" {
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
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{"id":"p1","name":"Policy 1","description":"desc","organization_id":"org","current_version_id":"v1","created_at":"2026-01-01","updated_at":"2026-01-01"}]}`))
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

	resp, err := client.Policies.List(1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].ID != "p1" {
		t.Fatalf("expected id=p1, got %s", resp.Results[0].ID)
	}
}

func TestPoliciesAPIRetrieve(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/policies/p1/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p1","name":"Policy 1","description":"desc","organization_id":"org","current_version_id":"v1","created_at":"2026-01-01","updated_at":"2026-01-01"}`))
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

	policy, err := client.Policies.Retrieve("p1")
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if policy.Name != "Policy 1" {
		t.Fatalf("expected name=Policy 1, got %s", policy.Name)
	}
}

func TestPoliciesAPICreate(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "Test Policy" {
			t.Fatalf("expected name=Test Policy, got %v", body["name"])
		}
		if body["description"] != "A test policy" {
			t.Fatalf("expected description, got %v", body["description"])
		}
		if _, ok := body["version_name"]; !ok {
			t.Fatalf("expected version_name in body")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p2","name":"Test Policy","description":"A test policy","organization_id":"org","current_version_id":"v1","created_at":"2026-01-01","updated_at":"2026-01-01"}`))
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

	policy, err := client.Policies.Create("Test Policy", map[string]any{"guidelines": "be safe"}, "A test policy", "v1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if policy.ID != "p2" {
		t.Fatalf("expected id=p2, got %s", policy.ID)
	}
}

func TestPoliciesAPIUpdate(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "Updated" {
			t.Fatalf("expected name=Updated, got %v", body["name"])
		}
		if _, ok := body["description"]; ok {
			t.Fatalf("did not expect description in body (was nil)")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p1","name":"Updated","description":"desc","organization_id":"org","current_version_id":"v1","created_at":"2026-01-01","updated_at":"2026-01-01"}`))
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

	name := "Updated"
	policy, err := client.Policies.Update("p1", &name, nil)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if policy.Name != "Updated" {
		t.Fatalf("expected name=Updated, got %s", policy.Name)
	}
}

func TestPoliciesAPIDelete(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/policies/p1/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
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

	err = client.Policies.Delete("p1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestPolicyVersionsListPaginated(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/policies/p1/versions/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"v1","version_name":"version 1","content":{"guidelines":"test"},"created_at":"2026-01-01","updated_at":"2026-01-01"}]}`))
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

	versions, err := client.Policies.Versions.List("p1")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].VersionName != "version 1" {
		t.Fatalf("expected version_name=version 1, got %s", versions[0].VersionName)
	}
}

func TestPolicyVersionsListRawArray(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"v1","version_name":"version 1","content":{},"created_at":"2026-01-01","updated_at":"2026-01-01"}]`))
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

	versions, err := client.Policies.Versions.List("p1")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}

func TestPolicyVersionsCreateRefetches(t *testing.T) {
	callCount := 0
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			// POST returns partial data
			_, _ = w.Write([]byte(`{"id":"v2"}`))
			return
		}
		// GET re-fetch returns full data
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/v2/") {
			_, _ = w.Write([]byte(`{"id":"v2","version_name":"v2","content":{"rules":"updated"},"created_at":"2026-01-01","updated_at":"2026-01-01"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
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

	version, err := client.Policies.Versions.Create("p1", map[string]any{"rules": "updated"}, "v2", "")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if version.ID != "v2" {
		t.Fatalf("expected id=v2, got %s", version.ID)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls (POST + GET), got %d", callCount)
	}
}

func TestCancelAndCancelAll(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
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

	if err := client.Agents.Jobs.Cancel("job-123"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := client.Agents.Jobs.CancelAll("agent-123"); err != nil {
		t.Fatalf("cancel all: %v", err)
	}
}

func TestMetadataOnRun(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		metaVal := r.FormValue("metadata")
		if metaVal == "" {
			t.Fatalf("expected metadata form field to be set")
		}
		var meta map[string]any
		if err := json.Unmarshal([]byte(metaVal), &meta); err != nil {
			t.Fatalf("failed to parse metadata JSON: %v", err)
		}
		if meta["source"] != "go-sdk-test" {
			t.Fatalf("expected source=go-sdk-test, got %v", meta["source"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"job-123"`))
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

	_, err = client.Agents.Run("agent-id", 0, map[string]any{
		"text": "hello",
	}, map[string]any{
		"source": "go-sdk-test",
	})
	if err != nil {
		t.Fatalf("run with metadata: %v", err)
	}
}
