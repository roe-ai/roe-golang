//go:build ignore

package main

import (
	"context"
	"log"
	"os"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	job, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
		"text": "Do work with timeout",
	})
	if err != nil {
		log.Fatalf("run: %v", err)
	}

	if _, err := job.WaitContext(ctx, 2*time.Second, 0); err != nil {
		log.Fatalf("wait: %v", err)
	}
}
