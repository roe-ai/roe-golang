package roe

import (
	"os"
	"strings"
	"testing"
)

func TestUserAgentReflectsVERSIONFile(t *testing.T) {
	raw, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	want := "roe-golang/" + strings.TrimSpace(string(raw))
	if userAgent != want {
		t.Fatalf("userAgent = %q, want %q", userAgent, want)
	}
}

func TestHeadersIncludeUserAgent(t *testing.T) {
	auth := newAuth(Config{APIKey: "k"})
	headers := auth.Headers()
	if headers.Get("User-Agent") != userAgent {
		t.Fatalf("User-Agent header = %q, want %q", headers.Get("User-Agent"), userAgent)
	}
}
