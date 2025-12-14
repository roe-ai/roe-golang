package roe

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPostDynamicInputsWithFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "roe-upload-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello world"); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tmp.Close()

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart content type, got %s", r.Header.Get("Content-Type"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}
		seenFile := false
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("read part: %v", err)
			}
			defer part.Close()
			switch part.FormName() {
			case "text":
				content, _ := io.ReadAll(part)
				if string(content) != "greeting" {
					t.Fatalf("unexpected text field %s", string(content))
				}
			case "upload":
				seenFile = true
				if part.FileName() == "" {
					t.Fatalf("expected filename on upload")
				}
				content, _ := io.ReadAll(part)
				if string(content) != "hello world" {
					t.Fatalf("unexpected file content: %s", string(content))
				}
			}
		}
		if !seenFile {
			t.Fatalf("expected to see file upload")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 10 * time.Millisecond,
		RetryMaxInterval:     10 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	var out map[string]bool
	err = client.postDynamicInputs("/upload", map[string]any{
		"text":   "greeting",
		"upload": FileUpload{Path: tmp.Name()},
	}, nil, &out)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if !out["ok"] {
		t.Fatalf("unexpected response: %v", out)
	}
}

func TestPostDynamicInputsWithURLInput(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Fatalf("expected urlencoded content type, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "upload=https%3A%2F%2Fexample.com%2Ffile.pdf" {
			t.Fatalf("unexpected form body: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 5 * time.Millisecond,
		RetryMaxInterval:     5 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	var out map[string]bool
	err := client.postDynamicInputs("/upload", map[string]any{
		"upload": FileUpload{URL: "https://example.com/file.pdf"},
	}, nil, &out)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
}
