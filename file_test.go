package roe

import (
	"io"
	"mime"
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
	}, nil, &out, nil)
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
	}, nil, &out, nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
}

func TestPostDynamicInputsEncodesSlicesAndMapsAsJSON(t *testing.T) {
	want := map[string]string{
		"list": `["a","b"]`,
		"obj":  `{"k":1}`,
		"flag": "true",
		"n":    "3",
	}
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if len(r.PostForm) != len(want) {
			t.Errorf("unexpected form fields: %v", r.PostForm)
		}
		for k, v := range want {
			if got := r.PostForm.Get(k); got != v {
				t.Errorf("%s = %q, want %q", k, got, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	cfg := Config{APIKey: "k", OrganizationID: "org", BaseURL: server.URL, Timeout: time.Second}
	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	inputs := map[string]any{"list": []string{"a", "b"}, "obj": map[string]any{"k": 1}, "flag": true, "n": 3, "none": nil}
	if err := client.postDynamicInputs("/upload", inputs, nil, nil, nil); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if err := client.postDynamicInputs("/upload", map[string]any{"bad": []any{make(chan int)}}, nil, nil, nil); err == nil {
		t.Fatalf("expected marshal error")
	}
}

func TestPrepareMultipartFileKeepsKnownMimeType(t *testing.T) {
	cases := []struct {
		file FileUpload
		want string
	}{
		{FileUpload{Reader: strings.NewReader("a,b\n1,2\n"), Filename: "data.csv"}, mime.TypeByExtension(".csv")},
		{FileUpload{Reader: strings.NewReader("{}"), MimeType: "application/json"}, "application/json"},
	}
	for _, tc := range cases {
		rc, _, got, err := (&httpClient{}).prepareMultipartFile(tc.file)
		if err != nil {
			t.Fatalf("prepareMultipartFile: %v", err)
		}
		rc.Close()
		if got != tc.want {
			t.Fatalf("mime type = %q, want %q", got, tc.want)
		}
	}
}

type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func TestPrepareMultipartFileClosesSniffedReader(t *testing.T) {
	src := &closeTracker{Reader: strings.NewReader("hello")}
	rc, _, _, err := (&httpClient{}).prepareMultipartFile(FileUpload{Reader: src})
	if err != nil {
		t.Fatalf("prepareMultipartFile: %v", err)
	}
	rc.Close()
	if !src.closed {
		t.Fatalf("expected the caller's reader to be closed")
	}
}

type closeCounter struct {
	io.Reader
	closes int
}

func (c *closeCounter) Close() error {
	c.closes++
	return nil
}

func TestMultipartUploadClosesEachReaderOnceWhenALaterFileFails(t *testing.T) {
	// Map order decides which file goes first, so repeat until the good
	// reader has been copied before the missing path fails.
	for i := 0; i < 20; i++ {
		src := &closeCounter{Reader: strings.NewReader("hello")}
		err := (&httpClient{}).postDynamicInputs("/upload", map[string]any{
			"good":    FileUpload{Reader: src, Filename: "a.txt"},
			"missing": FileUpload{Path: "/nonexistent/roe-upload.txt"},
		}, nil, nil, nil)
		if err == nil {
			t.Fatalf("expected an error for the missing file")
		}
		if src.closes > 1 {
			t.Fatalf("caller's reader closed %d times, want at most 1", src.closes)
		}
	}
}
