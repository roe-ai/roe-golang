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
	agent, err := client.Agents.CreateWithContext(
		ctx,
		"Example Agent",
		"MultimodalExtractionEngine",
		[]map[string]any{
			{"key": "text", "data_type": "text/plain", "description": "Text input"},
		},
		map[string]any{
			"model": "gpt-4.1-2025-04-14",
			"text":  "${text}",
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"summary": map[string]any{"type": "string"},
				},
			},
		},
		"", "",
	)
	if err != nil {
		log.Fatalf("create agent: %v", err)
	}
	fmt.Printf("Created agent %s\n", agent.ID)
}
