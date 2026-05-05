package roe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRawClientCallsGeneratedEndpoint(t *testing.T) {
	t.Helper()

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

	response, err := raw.V1UsersCurrentUserRetrieveWithResponse(context.Background())
	if err != nil {
		t.Fatalf("V1UsersCurrentUserRetrieveWithResponse returned error: %v", err)
	}

	if response.StatusCode() != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.StatusCode())
	}
	if authHeader != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %s", authHeader)
	}
}
