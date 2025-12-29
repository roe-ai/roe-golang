//go:build ignore

package main

import (
	"context"
	"fmt"
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

	batch, err := client.Agents.RunManyWithContext(ctx, "agent-uuid", []map[string]any{
		{"text": "Job 1"},
		{"text": "Job 2"},
		{"text": "Job 3"},
	}, 0)
	if err != nil {
		log.Fatalf("run many: %v", err)
	}

	results, err := batch.WaitContext(ctx, 0, 0)
	if err != nil {
		log.Fatalf("wait: %v", err)
	}
	fmt.Printf("Completed %d jobs\n", len(results))
}
