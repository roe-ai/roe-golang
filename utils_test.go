package roe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsUUIDString(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{
			name:  "ValidUUIDWithDashes",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  true,
		},
		{
			name:  "ValidUUIDWithoutDashes",
			input: "550e8400e29b41d4a716446655440000",
			want:  true,
		},
		{
			name:  "ValidUUIDUppercase",
			input: "550E8400-E29B-41D4-A716-446655440000",
			want:  true,
		},
		{
			name:  "ValidUUIDMixedCase",
			input: "550e8400-E29B-41d4-A716-446655440000",
			want:  true,
		},
		{
			name:  "EmptyString",
			input: "",
			want:  false,
		},
		{
			name:  "TooShort",
			input: "550e8400-e29b-41d4",
			want:  false,
		},
		{
			name:  "TooLong",
			input: "550e8400-e29b-41d4-a716-4466554400001234",
			want:  false,
		},
		{
			name:  "InvalidCharacters",
			input: "550e8400-e29b-41d4-a716-44665544zzzz",
			want:  false,
		},
		{
			name:  "RegularString",
			input: "hello-world",
			want:  false,
		},
		{
			name:  "FilePath",
			input: "/path/to/file.txt",
			want:  false,
		},
		{
			name:  "NonStringType",
			input: 12345,
			want:  false,
		},
		{
			name:  "NilValue",
			input: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUUIDString(tt.input)
			if got != tt.want {
				t.Errorf("isUUIDString(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLooksLikePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "UnixAbsolutePath",
			input: "/path/to/file.txt",
			want:  true,
		},
		{
			name:  "UnixRelativePath",
			input: "path/to/file.txt",
			want:  true,
		},
		{
			name:  "WindowsAbsolutePath",
			input: "C:\\Users\\test\\file.txt",
			want:  true,
		},
		{
			name:  "DotPrefix",
			input: "./relative/path",
			want:  true,
		},
		{
			name:  "DotDotPrefix",
			input: "../parent/path",
			want:  true,
		},
		{
			name:  "HiddenFile",
			input: ".hidden",
			want:  true,
		},
		{
			name:  "SimpleFilename",
			input: "file.txt",
			want:  false,
		},
		{
			name:  "UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  false,
		},
		{
			name:  "EmptyString",
			input: "",
			want:  false,
		},
		{
			name:  "JustText",
			input: "hello world",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikePath(tt.input)
			if got != tt.want {
				t.Errorf("looksLikePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsHTTPURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ValidHTTPS",
			input: "https://example.com",
			want:  true,
		},
		{
			name:  "ValidHTTP",
			input: "http://example.com",
			want:  true,
		},
		{
			name:  "HTTPSWithPath",
			input: "https://example.com/path/to/resource",
			want:  true,
		},
		{
			name:  "HTTPSWithQuery",
			input: "https://example.com/path?query=value",
			want:  true,
		},
		{
			name:  "HTTPSWithPort",
			input: "https://example.com:8080/path",
			want:  true,
		},
		{
			name:  "FTPScheme",
			input: "ftp://example.com",
			want:  false,
		},
		{
			name:  "FileScheme",
			input: "file:///path/to/file",
			want:  false,
		},
		{
			name:  "NoScheme",
			input: "example.com",
			want:  false,
		},
		{
			name:  "RelativePath",
			input: "/path/to/file",
			want:  false,
		},
		{
			name:  "EmptyString",
			input: "",
			want:  false,
		},
		{
			name:  "InvalidURL",
			input: "://invalid",
			want:  false,
		},
		{
			name:  "HTTPSNoHost",
			input: "https://",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHTTPURL(tt.input)
			if got != tt.want {
				t.Errorf("isHTTPURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsFilePath(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "ExistingFile",
			input: tmpFile,
			want:  true,
		},
		{
			name:  "ExistingDirectory",
			input: tmpDir,
			want:  false, // Directories should return false
		},
		{
			name:  "NonExistentFile",
			input: filepath.Join(tmpDir, "nonexistent.txt"),
			want:  false,
		},
		{
			name:  "EmptyString",
			input: "",
			want:  false,
		},
		{
			name:  "UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  false, // UUIDs should be excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFilePath(tt.input)
			if got != tt.want {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestChunkStrings(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		size     int
		wantLen  int
		wantLast int // Length of last chunk
	}{
		{
			name:     "ExactDivision",
			items:    []string{"a", "b", "c", "d"},
			size:     2,
			wantLen:  2,
			wantLast: 2,
		},
		{
			name:     "WithRemainder",
			items:    []string{"a", "b", "c", "d", "e"},
			size:     2,
			wantLen:  3,
			wantLast: 1,
		},
		{
			name:     "SingleChunk",
			items:    []string{"a", "b"},
			size:     10,
			wantLen:  1,
			wantLast: 2,
		},
		{
			name:     "EmptyInput",
			items:    []string{},
			size:     5,
			wantLen:  1,
			wantLast: 0,
		},
		{
			name:     "SizeZero",
			items:    []string{"a", "b"},
			size:     0,
			wantLen:  1,
			wantLast: 2,
		},
		{
			name:     "NegativeSize",
			items:    []string{"a", "b"},
			size:     -1,
			wantLen:  1,
			wantLast: 2,
		},
		{
			name:     "SizeEqualLength",
			items:    []string{"a", "b", "c"},
			size:     3,
			wantLen:  1,
			wantLast: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chunkStrings(tt.items, tt.size)
			if len(got) != tt.wantLen {
				t.Errorf("chunkStrings() returned %d chunks, want %d", len(got), tt.wantLen)
			}
			if len(got) > 0 && len(got[len(got)-1]) != tt.wantLast {
				t.Errorf("last chunk has %d items, want %d", len(got[len(got)-1]), tt.wantLast)
			}
		})
	}
}

func TestChunkAny(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		size     int
		wantLen  int
		wantLast int
	}{
		{
			name:     "ExactDivision",
			items:    []int{1, 2, 3, 4},
			size:     2,
			wantLen:  2,
			wantLast: 2,
		},
		{
			name:     "WithRemainder",
			items:    []int{1, 2, 3, 4, 5},
			size:     2,
			wantLen:  3,
			wantLast: 1,
		},
		{
			name:     "SingleChunk",
			items:    []int{1, 2},
			size:     10,
			wantLen:  1,
			wantLast: 2,
		},
		{
			name:     "EmptyInput",
			items:    []int{},
			size:     5,
			wantLen:  1,
			wantLast: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chunkAny(tt.items, tt.size)
			if len(got) != tt.wantLen {
				t.Errorf("chunkAny() returned %d chunks, want %d", len(got), tt.wantLen)
			}
			if len(got) > 0 && len(got[len(got)-1]) != tt.wantLast {
				t.Errorf("last chunk has %d items, want %d", len(got[len(got)-1]), tt.wantLast)
			}
		})
	}
}
