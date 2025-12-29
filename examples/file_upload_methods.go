//go:build ignore

package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/roe-ai/roe-golang"
)

func main() {
	client, err := roe.NewClient(
		os.Getenv("ROE_API_KEY"),
		os.Getenv("ROE_ORGANIZATION_ID"),
		os.Getenv("ROE_BASE_URL"),
		0,
		0,
	)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Path upload
	if _, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
		"pdf_files": "/path/to/file.pdf",
	}); err != nil {
		log.Printf("path upload: %v", err)
	}

	// FileUpload with override name
	if _, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
		"pdf_files": roe.FileUpload{
			Path:     "/path/to/file.pdf",
			Filename: "resume.pdf",
		},
	}); err != nil {
		log.Printf("file upload: %v", err)
	}

	// Reader upload
	reader := strings.NewReader("sample text")
	if _, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
		"text": reader,
	}); err != nil {
		log.Printf("reader upload: %v", err)
	}
}
