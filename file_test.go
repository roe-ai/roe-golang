package roe

import (
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strings"
	"testing"
	"time"
)

// TestDynamicInputsRequestWithFile verifies dynamicInputsRequest produces a
// well-formed multipart/form-data body with both scalar form fields and
// FileUpload parts (filename + content preserved).
func TestDynamicInputsRequestWithFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "roe-upload-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello world"); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	tmp.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 10 * time.Millisecond,
		RetryMaxInterval:     10 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}
	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	body, contentType, err := client.dynamicInputsRequest(map[string]any{
		"text":   "greeting",
		"upload": FileUpload{Path: tmp.Name()},
	}, nil)
	if err != nil {
		t.Fatalf("dynamicInputsRequest: %v", err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatalf("expected multipart content type, got %s", contentType)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("parse content type: %v / %s", err, mediaType)
	}
	reader := multipart.NewReader(body, params["boundary"])

	seenFile := false
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
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
		part.Close()
	}
	if !seenFile {
		t.Fatalf("expected to see file upload")
	}
}

// TestDynamicInputsRequestWithURLInput verifies that a FileUpload carrying
// only a URL collapses to a urlencoded form field — no multipart body, the
// URL stays intact as the field value.
func TestDynamicInputsRequestWithURLInput(t *testing.T) {
	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 5 * time.Millisecond,
		RetryMaxInterval:     5 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}
	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	body, contentType, err := client.dynamicInputsRequest(map[string]any{
		"upload": FileUpload{URL: "https://example.com/file.pdf"},
	}, nil)
	if err != nil {
		t.Fatalf("dynamicInputsRequest: %v", err)
	}
	if contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("expected urlencoded content type, got %s", contentType)
	}
	raw, _ := io.ReadAll(body)
	if string(raw) != "upload=https%3A%2F%2Fexample.com%2Ffile.pdf" {
		t.Fatalf("unexpected form body: %s", string(raw))
	}
}
