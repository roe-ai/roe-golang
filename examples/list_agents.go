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
	agents, err := client.Agents.ListWithContext(ctx, 1, 10)
	if err != nil {
		log.Fatalf("list: %v", err)
	}
	for _, agent := range agents.Results {
		fmt.Printf("%s: %s\n", agent.ID, agent.Name)
	}
}
