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

	outputs, err := client.Agents.RunSyncWithContext(ctx, "agent-uuid", map[string]any{
		"text": "Sync execution",
	})
	if err != nil {
		log.Fatalf("run sync: %v", err)
	}
	fmt.Printf("Received %d outputs\n", len(outputs))
}
