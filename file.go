package roe

import (
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
)

// FileUpload represents an explicit file upload with metadata.
// At least one of Path, Reader, or URL must be provided.
type FileUpload struct {
	Path     string
	Reader   io.Reader
	Filename string
	MimeType string
	URL      string

	// Optional validation; when set to >0, paths larger than this are rejected.
	MaxBytes int64
}

func (f FileUpload) isURL() bool {
	if f.URL == "" {
		return false
	}
	parsed, err := url.Parse(f.URL)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

// filename returns the effective filename.
func (f FileUpload) filename() string {
	if f.Filename != "" {
		return f.Filename
	}
	if f.Path != "" {
		return filepath.Base(f.Path)
	}
	if f.isURL() {
		return filepath.Base(f.URL)
	}
	return "upload"
}

// mimeType returns mime type or guesses from filename.
func (f FileUpload) mimeType() string {
	if f.MimeType != "" {
		return f.MimeType
	}
	if typ := mime.TypeByExtension(filepath.Ext(f.filename())); typ != "" {
		return typ
	}
	return "application/octet-stream"
}

// open returns an io.ReadCloser for the file upload.
func (f FileUpload) open() (io.ReadCloser, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	if f.Reader != nil {
		if rc, ok := f.Reader.(io.ReadCloser); ok {
			return rc, nil
		}
		return io.NopCloser(f.Reader), nil
	}

	if f.Path != "" {
		// Open file first to avoid TOCTOU race between Stat and Open
		file, err := os.Open(f.Path)
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", f.Path, err)
		}

		// Now stat the open file handle
		info, err := file.Stat()
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("stat file %s: %w", f.Path, err)
		}
		if info.IsDir() {
			file.Close()
			return nil, fmt.Errorf("file upload requires a file, got directory: %s", f.Path)
		}
		if info.Size() == 0 {
			file.Close()
			return nil, fmt.Errorf("file %s is empty", f.Path)
		}
		if f.MaxBytes > 0 && info.Size() > f.MaxBytes {
			file.Close()
			return nil, fmt.Errorf("file %s exceeds max size of %d bytes", f.Path, f.MaxBytes)
		}

		return file, nil
	}

	return nil, fmt.Errorf("file upload requires Path, Reader, or URL")
}

func (f FileUpload) validate() error {
	switch {
	case f.Reader != nil:
		return nil
	case f.Path != "":
		return nil
	case f.isURL():
		return nil
	default:
		return fmt.Errorf("file upload requires Path, Reader, or URL")
	}
}
