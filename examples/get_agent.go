//go:build ignore

package main

import (
	"context"
	"fmt"
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
	agent, err := client.Agents.RetrieveWithContext(ctx, "agent-uuid")
	if err != nil {
		log.Fatalf("retrieve: %v", err)
	}
	fmt.Printf("Agent: %s (%s)\n", agent.Name, agent.ID)
}
