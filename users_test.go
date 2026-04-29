package roe

import (
	"net/http"
	"testing"
	"time"
)

func TestUsersAPIMeReturnsCurrentUser(t *testing.T) {
	const userResponseJSON = `{"id":42,"email":"jane@example.com","first_name":"Jane","last_name":"Doe","display_name":"Jane Doe"}`

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/users/current_user/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got == "" {
			t.Fatalf("expected Authorization header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(userResponseJSON))
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

	user, err := client.Users.Me()
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if user == nil {
		t.Fatalf("expected non-nil user")
	}
	if string(user.Email) != "jane@example.com" {
		t.Fatalf("expected email=jane@example.com, got %s", user.Email)
	}
	if user.Id == nil || *user.Id != 42 {
		t.Fatalf("expected id=42, got %v", user.Id)
	}
	if user.FirstName == nil || *user.FirstName != "Jane" {
		t.Fatalf("expected first_name=Jane, got %v", user.FirstName)
	}
	if user.LastName == nil || *user.LastName != "Doe" {
		t.Fatalf("expected last_name=Doe, got %v", user.LastName)
	}
	if user.DisplayName == nil || *user.DisplayName != "Jane Doe" {
		t.Fatalf("expected display_name=Jane Doe, got %v", user.DisplayName)
	}
}

func TestUsersAPIMeSurfacesAPIError(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":"Invalid token"}`))
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

	if _, err := client.Users.Me(); err == nil {
		t.Fatalf("expected error from 401 response")
	}
}
