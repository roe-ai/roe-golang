// Package roe is the official Go SDK for the Roe AI API.
//
// # Installation
//
//	go get github.com/roe-ai/roe-golang@v1.0.0
//
// # Quick Start
//
// Create a client and run an agent:
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"os"
//		"time"
//
//		roe "github.com/roe-ai/roe-golang"
//	)
//
//	func main() {
//		client, err := roe.NewClient(
//			os.Getenv("ROE_API_KEY"),
//			os.Getenv("ROE_ORGANIZATION_ID"),
//			"", // baseURL (optional)
//			0,  // timeout (0 = default 60s)
//			0,  // retries (0 = default 3)
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer client.Close()
//
//		// Run an agent
//		job, err := client.Agents.Run("agent-uuid", 0, map[string]any{
//			"text": "Analyze this text",
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		result, err := job.Wait(5*time.Second, 0)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		for _, output := range result.Outputs {
//			fmt.Printf("%s: %s\n", output.Key, output.Value)
//		}
//	}
//
// # Core Features
//
//   - Agent management (list, create, retrieve, update, delete, duplicate)
//   - Agent version management
//   - Job execution (sync, async, batch)
//   - File upload support (path, URL, bytes, reader)
//   - Context-aware operations for cancellation support
//   - Automatic retry logic with exponential backoff
//   - Request/response hooks for monitoring
//   - Pagination support for list operations
//   - Reference file downloads (screenshots, HTML, markdown)
//
// # Environment Variables
//
//   - ROE_API_KEY: Your Roe AI API key
//   - ROE_ORGANIZATION_ID: Your organization UUID
//   - ROE_BASE_URL: Optional API base URL (defaults to https://api.roe-ai.com)
//   - ROE_TIMEOUT_SECONDS: Optional request timeout (defaults to 60s)
//   - ROE_MAX_RETRIES: Optional max retries (defaults to 3)
//
// # Links
//
//   - GitHub: https://github.com/roe-ai/roe-golang
//   - API Docs: https://docs.roe-ai.com
//   - Website: https://www.roe-ai.com/
package roe
