package roe

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	testOrgUUID         = "11111111-1111-1111-1111-111111111111"
	testPolicy1UUID     = "22222222-2222-2222-2222-222222222222"
	testPolicy2UUID     = "33333333-3333-3333-3333-333333333333"
	testVersion1UUID    = "44444444-4444-4444-4444-444444444444"
	testVersion2UUID    = "55555555-5555-5555-5555-555555555555"
	testCurrentVerUUID  = "66666666-6666-6666-6666-666666666666"
	policyResponseJSON1 = `{"id":"` + testPolicy1UUID + `","name":"Policy 1","description":"desc","organization_id":"` + testOrgUUID + `","current_version_id":"` + testCurrentVerUUID + `","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`
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
		if got := q.Get("organization_id"); got != testOrgUUID {
			t.Fatalf("expected organization_id=%s, got %q", testOrgUUID, got)
		}
		if got := q.Get("page"); got != "1" {
			t.Fatalf("expected page=1, got %q", got)
		}
		if got := q.Get("page_size"); got != "10" {
			t.Fatalf("expected page_size=10, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[` + policyResponseJSON1 + `]}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
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
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Id == nil || resp.Results[0].Id.String() != testPolicy1UUID {
		t.Fatalf("expected id=%s, got %v", testPolicy1UUID, resp.Results[0].Id)
	}
	if resp.Results[0].Name != "Policy 1" {
		t.Fatalf("expected name=Policy 1, got %s", resp.Results[0].Name)
	}
}

func TestPoliciesAPIRetrieve(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		expectedPath := "/v1/policies/" + testPolicy1UUID + "/"
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path %s, want %s", r.URL.Path, expectedPath)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(policyResponseJSON1))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	policy, err := client.Policies.Retrieve(testPolicy1UUID)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if policy == nil || policy.Name != "Policy 1" {
		t.Fatalf("expected name=Policy 1, got %v", policy)
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
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"` + testPolicy2UUID + `","name":"Test Policy","description":"A test policy","organization_id":"` + testOrgUUID + `","current_version_id":"` + testCurrentVerUUID + `","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
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
	if policy == nil || policy.Id == nil || policy.Id.String() != testPolicy2UUID {
		t.Fatalf("expected id=%s, got %v", testPolicy2UUID, policy)
	}
}

func TestPoliciesAPIUpdate(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
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
		_, _ = w.Write([]byte(`{"id":"` + testPolicy1UUID + `","name":"Updated","description":"desc","organization_id":"` + testOrgUUID + `","current_version_id":"` + testCurrentVerUUID + `","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	name := "Updated"
	policy, err := client.Policies.Update(testPolicy1UUID, &name, nil)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if policy == nil || policy.Name != "Updated" {
		t.Fatalf("expected name=Updated, got %v", policy)
	}
}

func TestPoliciesAPIDelete(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		expectedPath := "/v1/policies/" + testPolicy1UUID + "/"
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if err := client.Policies.Delete(testPolicy1UUID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestPolicyVersionsListPaginated(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/policies/"+testPolicy1UUID+"/versions/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{"id":"` + testVersion1UUID + `","version_name":"version 1","content":{"guidelines":"test"},"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","policy":` + policyResponseJSON1 + `}]}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	versions, err := client.Policies.Versions.List(testPolicy1UUID)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if versions == nil || len(versions.Results) != 1 {
		t.Fatalf("expected 1 version, got %v", versions)
	}
	if versions.Results[0].VersionName != "version 1" {
		t.Fatalf("expected version_name=version 1, got %s", versions.Results[0].VersionName)
	}
}

func TestPolicyVersionsCreate(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"` + testVersion2UUID + `","version_name":"v2","content":{"rules":"updated"}}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	version, err := client.Policies.Versions.Create(testPolicy1UUID, map[string]any{"rules": "updated"}, "v2", "")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if version == nil || version.Id == nil || version.Id.String() != testVersion2UUID {
		t.Fatalf("expected id=%s, got %v", testVersion2UUID, version)
	}
}

// Agent-side tests retained here from the original policies_test.go because
// they exercise the still-hand-written agents.go path.
func TestCancelVerifiesPath(t *testing.T) {
	const jobID = "11112222-3333-4444-5555-666677778888"
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		expected := "/v1/agents/jobs/" + jobID + "/cancel/"
		if r.URL.Path != expected {
			t.Fatalf("expected cancel path %s, got %s", expected, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if err := client.Agents.Jobs.Cancel(jobID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
}

func TestCancelAllVerifiesPath(t *testing.T) {
	const agentID = "99998888-7777-6666-5555-444433332222"
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		expected := "/v1/agents/" + agentID + "/jobs/cancel-all/"
		if r.URL.Path != expected {
			t.Fatalf("expected cancel-all path %s, got %s", expected, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if err := client.Agents.Jobs.CancelAll(agentID); err != nil {
		t.Fatalf("cancel all: %v", err)
	}
}

func TestMetadataOnRun(t *testing.T) {
	const agentID = "abcdef12-3456-7890-abcd-ef1234567890"
	const jobID = "11112222-3333-4444-5555-666677778888"
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
		_, _ = w.Write([]byte(`"` + jobID + `"`))
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	_, err = client.Agents.Run(agentID, 0, map[string]any{
		"text": "hello",
	}, map[string]any{
		"source": "go-sdk-test",
	})
	if err != nil {
		t.Fatalf("run with metadata: %v", err)
	}
}

func TestMetadataKeyCollisionReturnsError(t *testing.T) {
	const agentID = "abcdef12-3456-7890-abcd-ef1234567890"
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not reach server")
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: testOrgUUID,
		BaseURL:        server.URL,
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	_, err = client.Agents.Run(agentID, 0, map[string]any{
		"text":     "hello",
		"metadata": "should-collide",
	}, map[string]any{
		"source": "go-sdk-test",
	})
	if err == nil {
		t.Fatalf("expected error for metadata key collision")
	}
	if !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("expected error about metadata collision, got: %v", err)
	}
}
