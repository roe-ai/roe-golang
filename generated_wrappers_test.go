package roe

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGeneratedAPIsAreExposedOnClient(t *testing.T) {
	client, err := NewClientWithConfig(Config{
		APIKey:         "k",
		OrganizationID: "11111111-1111-1111-1111-111111111111",
		BaseURL:        "https://example.com",
		Timeout:        time.Second,
		MaxRetries:     0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if client.Discovery == nil {
		t.Fatalf("expected Discovery API to be initialized")
	}
	if client.Tables == nil {
		t.Fatalf("expected Tables API to be initialized")
	}
}

func TestDiscoveryAPIListSupportedModelsSendsCapability(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/models/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("capability"); got != "text" {
			t.Fatalf("expected capability=text, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[],"tenant_scope":"all","total_count":0}`))
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

	resp, err := client.Discovery.ListSupportedModels("text")
	if err != nil {
		t.Fatalf("list supported models: %v", err)
	}
	if resp.TotalCount != 0 {
		t.Fatalf("unexpected total count: %d", resp.TotalCount)
	}
}

func TestTablesAPIUploadPostsMultipartFields(t *testing.T) {
	tmp, err := os.CreateTemp("", "roe-table-*.csv")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("name\nAda\n"); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tables/upload/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart content type, got %s", r.Header.Get("Content-Type"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}

		seen := map[string]string{}
		seenFile := false
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			content, _ := io.ReadAll(part)
			switch part.FormName() {
			case "file":
				seenFile = true
				if part.FileName() == "" {
					t.Fatalf("expected filename on file part")
				}
				if string(content) != "name\nAda\n" {
					t.Fatalf("unexpected file content %q", string(content))
				}
			default:
				seen[part.FormName()] = string(content)
			}
		}
		if !seenFile {
			t.Fatalf("expected file part")
		}
		if seen["table_name"] != "customers" || seen["with_headers"] != "true" {
			t.Fatalf("unexpected form fields: %v", seen)
		}
		if seen["organization_id"] != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("unexpected organization_id: %q", seen["organization_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organization_id":"11111111-1111-1111-1111-111111111111","table_name":"customers","summary":{}}`))
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

	resp, err := client.Tables.Upload("customers", FileUpload{Path: tmp.Name()}, true)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if resp.TableName != "customers" {
		t.Fatalf("expected table customers, got %s", resp.TableName)
	}
}
