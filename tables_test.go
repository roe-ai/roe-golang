package roe

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTablesAPIUploadSendsMultipart(t *testing.T) {
	var capturedTable, capturedHeaders, capturedOrg string
	var capturedFileName string
	var capturedFileBody []byte

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tables/upload/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart Content-Type, got %s", r.Header.Get("Content-Type"))
		}

		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("parse content type: %v", err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			data, _ := io.ReadAll(part)
			switch part.FormName() {
			case "table_name":
				capturedTable = string(data)
			case "with_headers":
				capturedHeaders = string(data)
			case "organization_id":
				capturedOrg = string(data)
			case "file":
				capturedFileName = part.FileName()
				capturedFileBody = data
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"table_name":"customers","organization_id":"org","summary":{"written_rows":1}}`))
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

	result, err := client.Tables.Upload(
		"customers",
		FileUpload{Reader: bytes.NewReader([]byte("name\nAda\n")), Filename: "customers.csv"},
		nil,
	)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.TableName != "customers" {
		t.Fatalf("expected table_name=customers, got %s", result.TableName)
	}
	if capturedTable != "customers" {
		t.Fatalf("expected table_name multipart=customers, got %q", capturedTable)
	}
	if capturedHeaders != "true" {
		t.Fatalf("expected with_headers=true, got %q", capturedHeaders)
	}
	if capturedOrg != "org" {
		t.Fatalf("expected organization_id=org, got %q", capturedOrg)
	}
	if capturedFileName != "customers.csv" {
		t.Fatalf("expected filename customers.csv, got %q", capturedFileName)
	}
	if string(capturedFileBody) != "name\nAda\n" {
		t.Fatalf("unexpected file body %q", capturedFileBody)
	}
}

func TestTablesAPIUploadWithHeadersFalse(t *testing.T) {
	var captured string
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			if part.FormName() == "with_headers" {
				data, _ := io.ReadAll(part)
				captured = string(data)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"table_name":"t","organization_id":"org","summary":{}}`))
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

	falseVal := false
	if _, err := client.Tables.Upload(
		"t",
		FileUpload{Reader: bytes.NewReader([]byte("x\n1\n")), Filename: "t.csv"},
		&TableUploadOptions{WithHeaders: &falseVal},
	); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if captured != "false" {
		t.Fatalf("expected with_headers=false, got %q", captured)
	}
}
