package roe

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestConnectionsAPIUpdateAndReplaceUseDistinctHTTPMethods(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		call       func(*RoeClient) error
	}{
		{
			name:       "update",
			wantMethod: http.MethodPatch,
			call: func(client *RoeClient) error {
				_, err := client.Connections.Update("conn_1", "Updated", "desc", map[string]any{"database": "analytics"}, nil)
				return err
			},
		},
		{
			name:       "replace",
			wantMethod: http.MethodPut,
			call: func(client *RoeClient) error {
				_, err := client.Connections.Replace("conn_1", "Updated", "desc", map[string]any{"database": "analytics"}, nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod {
					t.Fatalf("expected %s, got %s", tt.wantMethod, r.Method)
				}
				if r.URL.Path != "/v1/connections/conn_1/" {
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("organization_id"); got != "11111111-1111-1111-1111-111111111111" {
					t.Fatalf("expected organization_id query, got %q", got)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body["name"] != "Updated" {
					t.Fatalf("expected name=Updated, got %v", body["name"])
				}
				if body["description"] != "desc" {
					t.Fatalf("expected description=desc, got %v", body["description"])
				}
				config, ok := body["config"].(map[string]any)
				if !ok {
					t.Fatalf("expected config object, got %T", body["config"])
				}
				if config["database"] != "analytics" {
					t.Fatalf("expected config database=analytics, got %v", config["database"])
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"22222222-2222-2222-2222-222222222222","connector_type":"snowflake","name":"Updated","organization":"11111111-1111-1111-1111-111111111111"}`))
			}))
			defer server.Close()

			client, err := NewClientWithConfig(Config{
				APIKey:         "k",
				OrganizationID: "11111111-1111-1111-1111-111111111111",
				BaseURL:        server.URL,
				Timeout:        time.Second,
				MaxRetries:     0,
			})
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			defer client.Close()

			if err := tt.call(client); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
		})
	}
}
