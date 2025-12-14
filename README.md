# Roe AI Go SDK

A Go SDK for the [Roe AI](https://www.roe-ai.com/) API.

## Installation

Install the latest version:

```bash
go get github.com/roe-ai/roe-golanglang
```

Or pin to a specific version (e.g., v0.1.0):

```bash
go get github.com/roe-ai/roe-golanglang@v0.1.0
```

**Requirements:** Go 1.23+

**Note:** `github.com/roe-ai/roe-golanglang/roe` is a deprecated compatibility shim. Prefer importing `github.com/roe-ai/roe-golanglang` directly.

## Quick Start

```go
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
	// Initialize client with environment variables or explicit credentials
	client, err := roe.NewClient(
		os.Getenv("ROE_API_KEY"),
		os.Getenv("ROE_ORGANIZATION_ID"),
		"", // baseURL (optional)
		0,  // timeout in seconds (0 uses default 60s)
		0,  // max retries (0 uses default 3)
	)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run an agent asynchronously
	job, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
		"text": "Analyze this text",
	})
	if err != nil {
		log.Fatalf("failed to run agent: %v", err)
	}

	// Wait for job to complete
	result, err := job.WaitContext(ctx, 2*time.Second, 0)
	if err != nil {
		log.Fatalf("job failed: %v", err)
	}

	// Process outputs
	for _, output := range result.Outputs {
		fmt.Printf("%s: %s\n", output.Key, output.Value)
	}
}
```

Set environment variables:

```bash
export ROE_API_KEY="your-api-key"
export ROE_ORGANIZATION_ID="your-org-uuid"
```

## Configuration

- Defaults: 60s timeout, 3 retries, automatic request IDs, and connection pooling. Enable verbose request/response logs with `ROE_DEBUG=true` (sensitive headers are redacted).
- Env vars: `ROE_TIMEOUT`, `ROE_MAX_RETRIES`, `ROE_DEBUG`, `ROE_PROXY`, `ROE_EXTRA_HEADERS`, `ROE_REQUEST_ID`, `ROE_AUTO_REQUEST_ID`, `ROE_REQUEST_ID_HEADER`, `ROE_RETRY_INITIAL_MS`, `ROE_RETRY_MAX_MS`, `ROE_RETRY_MULTIPLIER`, `ROE_RETRY_JITTER`, `ROE_MAX_IDLE_CONNS`, `ROE_MAX_IDLE_CONNS_PER_HOST`, `ROE_IDLE_CONN_TIMEOUT`.
- Programmatic options via `roe.ConfigParams`:

```go
debug := true
client, err := roe.NewClientWithParams(roe.ConfigParams{
  APIKey:         os.Getenv("ROE_API_KEY"),
  OrganizationID: os.Getenv("ROE_ORGANIZATION_ID"),
  Timeout:        30 * time.Second,
  MaxRetries:     5,
  Debug:          &debug,
  ProxyURL:       "http://localhost:8080",
  ExtraHeaders:   http.Header{"X-Request-ID": []string{"my-request"}},
})
```

## Error handling & retries

- Typed errors: `BadRequestError`, `AuthenticationError`, `InsufficientCreditsError`, `ForbiddenError`, `NotFoundError`, `RateLimitError` (with `RetryAfter`), `ServerError`, all embedding `APIError` with `RequestID`, status, parsed detail, and raw body.
- Retries with configurable backoff/jitter for 5xx/429/408 and network errors. Request/response logging (redacted) is available when `Debug` is enabled.

## Polling & batches

- `Job.WaitContext` and `JobBatch.WaitContext` accept contexts/timeouts, preserve input order, avoid redundant requests, and short-circuit on failure/cancellation with helpful errors.
- Batch status/result helpers keep ordering aligned with the input job IDs.

## File uploads

- `FileUpload` accepts a path, reader, byte slices, or a URL string. Missing/invalid paths return clear errors.
- Dynamic inputs also accept plain paths or UUIDs; URLs are sent as form values.
- Example:

```go
job, _ := client.Agents.Run("agent-id", 0, map[string]any{
  "pdf_file": roe.FileUpload{Path: "/tmp/doc.pdf"},
  "metadata": "customer:123",
})
result, _ := job.Wait(0, 0)
```

## Parity snapshot

| Area | Go SDK status |
| --- | --- |
| Config/env parsing & proxies | ✅ (`ROE_*` envs, proxy + extra headers, request IDs) |
| HTTP client | ✅ retries with backoff/jitter, request/response logging w/ redaction |
| Error taxonomy | ✅ typed errors incl. rate limits + request IDs |
| Batches/polling | ✅ order-preserving `WaitContext`, early failure surfacing |
| File inputs | ✅ paths, readers/bytes, URLs, validation |

## Agent Examples

### Multimodal Extraction

```go
agent, _ := client.Agents.Create(
	"Listing Analyzer",
	"MultimodalExtractionEngine",
	[]map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Item description"},
	},
	map[string]any{
		"model": "gpt-4.1-2025-04-14",
		"text": "${text}",
		"instruction": "Analyze this product listing. Is it counterfeit?",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"is_counterfeit": map[string]any{"type": "boolean", "description": "Whether likely counterfeit"},
				"confidence":     map[string]any{"type": "number", "description": "Confidence score 0-1"},
				"reasoning":      map[string]any{"type": "string", "description": "Explanation"},
			},
		},
	},
	"", "",
)

job, _ := client.Agents.Run(agent.ID, 0, map[string]any{
	"text": "Authentic Louis Vuitton bag, brand new, $50",
})
result, _ := job.Wait(0, 0)
```

### Document Insights

```go
agent, _ := client.Agents.Create(
	"Resume Parser",
	"PDFExtractionEngine",
	[]map[string]any{
		{"key": "pdf_files", "data_type": "application/pdf", "description": "Resume PDF"},
	},
	map[string]any{
		"model": "gpt-4.1-2025-04-14",
		"pdf_files": "${pdf_files}",
		"instructions": "Extract candidate information from this resume.",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":  map[string]any{"type": "string"},
				"email": map[string]any{"type": "string"},
				"skills": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
			},
		},
	},
	"", "",
)

job, _ := client.Agents.Run(agent.ID, 0, map[string]any{
	"pdf_files": "/path/to/resume.pdf",
})
result, _ := job.Wait(0, 0)
```

### Web Insights

```go
agent, _ := client.Agents.Create(
	"Company Analyzer",
	"URLWebsiteExtractionEngine",
	[]map[string]any{
		{"key": "url", "data_type": "text/plain", "description": "Website URL"},
	},
	map[string]any{
		"url":    "${url}",
		"model":  "gpt-4.1-2025-04-14",
		"instruction": "Extract company information from this website.",
		"vision_mode": false,
		"crawl_config": map[string]any{
			"save_html":      true,
			"save_markdown":  true,
			"save_screenshot": true,
		},
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"company_name": map[string]any{"type": "string"},
				"description":  map[string]any{"type": "string"},
				"products": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
			},
		},
	},
	"", "",
)

job, _ := client.Agents.Run(agent.ID, 0, map[string]any{
	"url": "https://www.roe-ai.com/",
})
result, _ := job.Wait(0, 0)

for _, ref := range result.GetReferences() {
	content, _ := client.Agents.Jobs.DownloadReference(job.ID(), ref.ResourceID, false)
	_ = content // write to file as needed
}
```

## CI & tests

- Unit tests: `go test ./... -race`. Integration tests under `./tests` are skipped unless `ROE_API_KEY` and `ROE_ORGANIZATION_ID` are set.
- See `CHANGELOG.md` for recent parity updates.

### Interactive Web

```go
agent, _ := client.Agents.Create(
	"Meeting Booker",
	"InteractiveWebExtractionEngine",
	[]map[string]any{
		{"key": "url", "data_type": "text/plain", "description": "Website URL"},
		{"key": "action", "data_type": "text/plain", "description": "Action to perform"},
	},
	map[string]any{
		"url":    "${url}",
		"action": "${action}",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"calendar_link": map[string]any{"type": "string", "description": "Booking link found"},
				"steps_taken":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
		},
	},
	"", "",
)

job, _ := client.Agents.Run(agent.ID, 0, map[string]any{
	"url":    "https://www.roe-ai.com/",
	"action": "Find the founder's calendar link to book a meeting",
})
result, _ := job.Wait(0, 0)
```

## Running Agents

```go
// Async (recommended)
job, err := client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
	"text": "input",
})
if err != nil {
	log.Fatalf("run failed: %v", err)
}
result, err := job.WaitContext(ctx, 2*time.Second, 0)
if err != nil {
	log.Fatalf("wait failed: %v", err)
}

// Sync (blocking)
outputs, err := client.Agents.RunSyncWithContext(ctx, "agent-uuid", map[string]any{
	"text": "input",
})
if err != nil {
	log.Fatalf("run sync failed: %v", err)
}

// With files (auto-uploaded)
job, err = client.Agents.RunWithContext(ctx, "agent-uuid", 0, map[string]any{
	"document": roe.FileUpload{Path: "/path/to/file.pdf"},
})
if err != nil {
	log.Fatalf("run with file failed: %v", err)
}

// Batch processing
batch, err := client.Agents.RunManyWithContext(ctx, "agent-uuid", []map[string]any{
	{"text": "input1"},
	{"text": "input2"},
	{"text": "input3"},
}, 0)
if err != nil {
	log.Fatalf("batch run failed: %v", err)
}
results, err := batch.WaitContext(ctx, 2*time.Second, 0)
if err != nil {
	log.Fatalf("batch wait failed: %v", err)
}
```

## Agent Management

```go
// List agents
agents, err := client.Agents.ListWithContext(ctx, 1, 10)
if err != nil {
	log.Fatalf("list failed: %v", err)
}

// Retrieve agent
agent, err := client.Agents.RetrieveWithContext(ctx, "agent-uuid")
if err != nil {
	log.Fatalf("retrieve failed: %v", err)
}

// Update agent
updated, err := client.Agents.UpdateWithContext(ctx, "agent-uuid", "New Name", nil, nil)
if err != nil {
	log.Fatalf("update failed: %v", err)
}

// Delete agent
err = client.Agents.DeleteWithContext(ctx, "agent-uuid")
if err != nil {
	log.Fatalf("delete failed: %v", err)
}

// Duplicate agent
duplicate, err := client.Agents.DuplicateWithContext(ctx, "agent-uuid")
if err != nil {
	log.Fatalf("duplicate failed: %v", err)
}
```

## Version Management

```go
// List versions
versions, err := client.Agents.Versions.ListWithContext(ctx, "agent-uuid")
if err != nil {
	log.Fatalf("list versions failed: %v", err)
}

// Get current version
current, err := client.Agents.Versions.RetrieveCurrentWithContext(ctx, "agent-uuid")
if err != nil {
	log.Fatalf("get current version failed: %v", err)
}

// Create version
version, err := client.Agents.Versions.CreateWithContext(
	ctx,
	"agent-uuid",
	[]map[string]any{{"key": "text", "data_type": "text/plain", "description": "Text input"}},
	map[string]any{"model": "gpt-4.1-2025-04-14"},
	"v2",
	"",
)
if err != nil {
	log.Fatalf("create version failed: %v", err)
}

// Update version
err = client.Agents.Versions.UpdateWithContext(ctx, "agent-uuid", version.ID, "v2-updated", "")
if err != nil {
	log.Fatalf("update version failed: %v", err)
}

// Delete version
err = client.Agents.Versions.DeleteWithContext(ctx, "agent-uuid", version.ID)
if err != nil {
	log.Fatalf("delete version failed: %v", err)
}

// Run specific version
job, err := client.Agents.RunVersionWithContext(ctx, "agent-uuid", version.ID, 0, map[string]any{"text": "input"})
if err != nil {
	log.Fatalf("run version failed: %v", err)
}
result, err := job.WaitContext(ctx, 2*time.Second, 0)
if err != nil {
	log.Fatalf("wait failed: %v", err)
}
```

## Job Management

```go
// Retrieve job status
status, err := client.Agents.Jobs.RetrieveStatusWithContext(ctx, "job-id")
if err != nil {
	log.Fatalf("get status failed: %v", err)
}

// Retrieve job result
result, err := client.Agents.Jobs.RetrieveResultWithContext(ctx, "job-id")
if err != nil {
	log.Fatalf("get result failed: %v", err)
}

// Retrieve multiple job statuses
statuses, err := client.Agents.Jobs.RetrieveStatusManyWithContext(ctx, []string{"job1", "job2"})
if err != nil {
	log.Fatalf("get statuses failed: %v", err)
}

// Retrieve multiple job results
results, err := client.Agents.Jobs.RetrieveResultManyWithContext(ctx, []string{"job1", "job2"})
if err != nil {
	log.Fatalf("get results failed: %v", err)
}

// Download reference file
content, err := client.Agents.Jobs.DownloadReferenceWithContext(ctx, "job-id", "resource-id", false)
if err != nil {
	log.Fatalf("download reference failed: %v", err)
}

// Delete job data
deleteResp, err := client.Agents.Jobs.DeleteDataWithContext(ctx, "job-id")
if err != nil {
	log.Fatalf("delete data failed: %v", err)
}
```

## Supported Models

| Model | Value |
|-------|-------|
| GPT-5.1 | `gpt-5.1-2025-11-13` |
| GPT-5 | `gpt-5-2025-08-07` |
| GPT-5 Mini | `gpt-5-mini-2025-08-07` |
| GPT-4.1 | `gpt-4.1-2025-04-14` |
| GPT-4.1 Mini | `gpt-4.1-mini-2025-04-14` |
| O3 Pro | `o3-pro-2025-06-10` |
| O3 | `o3-2025-04-16` |
| O4 Mini | `o4-mini-2025-04-16` |
| GPT-4o | `gpt-4o-2024-11-20` |
| Claude Sonnet 4.5 | `claude-sonnet-4-5-20250929` |
| Claude Sonnet 4 | `claude-sonnet-4-20250514` |
| Claude 3.7 Sonnet | `claude-3-7-sonnet-20250219` |
| Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| Claude 3.5 Haiku | `claude-3-5-haiku-20241022` |
| Claude Opus 4.5 | `claude-opus-4-5-20251101` |
| Claude Opus 4.1 | `claude-opus-4-1-20250805` |
| Claude Opus 4 | `claude-opus-4-20250514` |
| Gemini 3 Pro | `gemini-3-pro-preview` |
| Gemini 2.5 Pro | `gemini-2.5-pro` |
| Gemini 2.5 Flash | `gemini-2.5-flash` |

## Engine Classes

| Engine | ID |
|--------|----|
| Multimodal Extraction | `MultimodalExtractionEngine` |
| Document Insights | `PDFExtractionEngine` |
| Document Segmentation | `PDFPageSelectionEngine` |
| Web Insights | `URLWebsiteExtractionEngine` |
| Interactive Web | `InteractiveWebExtractionEngine` |
| Web Search | `URLFinderEngine` |
| Perplexity Search | `PerplexitySearchEngine` |
| Maps Search | `GoogleMapsEntityExtractionEngine` |
| Merchant Risk | `MerchantRiskAnalysisEngine` |
| Product Policy | `ProductPolicyEngine` |
| LinkedIn Crawler | `LinkedInScraperEngine` |
| Social Media | `SocialScraperEngine` |

## Links

- [Roe AI](https://www.roe-ai.com/)
- [API Documentation](https://docs.roe-ai.com)
- [Examples](examples/)
