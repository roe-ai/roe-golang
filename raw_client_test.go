package roe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/roe-ai/roe-golang/generated"
)

// Verifies client.Raw() returns a generated client wired with the same base
// URL, http.Doer, and auth editor as the ergonomic SDK surface. The test goes
// through the generic *generated.Client plumbing rather than a specific
// operation function so it survives codegen renames driven by upstream
// OpenAPI operationId changes (e.g. the v1.0.79 V1*-prefix removal).
func TestRawClientCallsGeneratedEndpoint(t *testing.T) {
	var authHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/users/current_user/" {
			t.Fatalf("unexpected raw path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient("test-key", "test-org", server.URL, 0, 0)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	raw, err := client.Raw()
	if err != nil {
		t.Fatalf("Raw returned error: %v", err)
	}

	gen, ok := raw.ClientInterface.(*generated.Client)
	if !ok {
		t.Fatalf("Raw().ClientInterface is %T, want *generated.Client", raw.ClientInterface)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, gen.Server+"v1/users/current_user/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	for _, edit := range gen.RequestEditors {
		if err := edit(context.Background(), req); err != nil {
			t.Fatalf("RequestEditor returned error: %v", err)
		}
	}

	resp, err := gen.Client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if authHeader != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %s", authHeader)
	}
}
