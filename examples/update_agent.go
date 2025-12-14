//go:build ignore

package main

import (
	"context"
	"log"
	"os"

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
	disableCache := true
	cacheFailed := true
	if _, err := client.Agents.UpdateWithContext(ctx, "agent-uuid", "Updated Name", &disableCache, &cacheFailed); err != nil {
		log.Fatalf("update: %v", err)
	}
	if err := client.Agents.DeleteWithContext(ctx, "agent-uuid"); err != nil {
		log.Printf("cleanup: %v", err)
	}
}
