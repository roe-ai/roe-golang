# Roe AI Go SDK

Go SDK for the [Roe AI](https://www.roe-ai.com/) API.

## Installation

```bash
go get github.com/roe-ai/roe-golang
```

Requires Go 1.23+

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"

    roe "github.com/roe-ai/roe-golang"
)

func main() {
    client, err := roe.NewClient(
        os.Getenv("ROE_API_KEY"),
        os.Getenv("ROE_ORGANIZATION_ID"),
        "", // baseURL (optional)
        0,  // timeout (0 = default 60s)
        0,  // retries (0 = default 3)
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Run an agent
    job, err := client.Agents.Run("agent-uuid", 0, map[string]any{
        "text": "Analyze this text",
    })
    if err != nil {
        log.Fatal(err)
    }

    result, err := job.Wait(5*time.Second, 0)
    if err != nil {
        log.Fatal(err)
    }

    for _, output := range result.Outputs {
        fmt.Printf("%s: %s\n", output.Key, output.Value)
    }
}
```

Or use environment variables:

```bash
export ROE_API_KEY="your-api-key"
export ROE_ORGANIZATION_ID="your-org-uuid"
```

## Full Example

Create an agent that extracts structured data from websites:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    roe "github.com/roe-ai/roe-golang"
)

func main() {
    client, _ := roe.NewClient(
        os.Getenv("ROE_API_KEY"),
        os.Getenv("ROE_ORGANIZATION_ID"),
        "", 0, 0,
    )
    defer client.Close()

    // Create a Web Insights agent
    agent, _ := client.Agents.Create(
        "Company Analyzer",
        "URLWebsiteExtractionEngine",
        []map[string]any{
            {"key": "url", "data_type": "text/plain", "description": "Website URL"},
        },
        map[string]any{
            "url":         "${url}",
            "model":       "gpt-4.1-2025-04-14",
            "instruction": "Extract company information from this website.",
            "vision_mode": false,
            "crawl_config": map[string]any{
                "save_html":       true,
                "save_markdown":   true,
                "save_screenshot": true,
            },
            "output_schema": map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "company_name": map[string]any{"type": "string"},
                    "description":  map[string]any{"type": "string"},
                    "products":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
                },
            },
        },
        "", "",
    )

    // Run the agent
    job, _ := client.Agents.Run(agent.ID, 0, map[string]any{
        "url": "https://www.roe-ai.com/",
    })
    result, _ := job.Wait(5*time.Second, 0)

    // Print results
    for _, output := range result.Outputs {
        var parsed map[string]any
        json.Unmarshal([]byte(output.Value), &parsed)
        prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
        fmt.Println(string(prettyJSON))
    }

    // Download saved references (screenshots, HTML, markdown)
    for _, ref := range result.GetReferences() {
        content, _ := client.Agents.Jobs.DownloadReference(job.ID(), ref.ResourceID, false)
        os.WriteFile(ref.ResourceID, content, 0644)
    }

    // Cleanup
    client.Agents.Delete(agent.ID)
}
```

## API Reference

### Agents

```go
client.Agents.List(page, pageSize)
client.Agents.Retrieve(agentID)
client.Agents.Create(name, engineClassID, inputDefs, engineConfig, versionName, desc)
client.Agents.Update(agentID, name, disableCache, cacheFailedJobs)
client.Agents.Delete(agentID)
client.Agents.Duplicate(agentID)
```

### Running Agents

```go
client.Agents.Run(agentID, timeout, inputs)
client.Agents.RunSync(agentID, inputs)
client.Agents.RunMany(agentID, batchInputs, timeout)
client.Agents.RunVersion(agentID, versionID, timeout, inputs)
```

### Versions

```go
client.Agents.Versions.List(agentID)
client.Agents.Versions.Retrieve(agentID, versionID)
client.Agents.Versions.RetrieveCurrent(agentID)
client.Agents.Versions.Create(agentID, inputDefs, engineConfig, versionName, desc)
client.Agents.Versions.Update(agentID, versionID, versionName, desc)
client.Agents.Versions.Delete(agentID, versionID)
```

### Jobs

```go
client.Agents.Jobs.RetrieveStatus(jobID)
client.Agents.Jobs.RetrieveResult(jobID)
client.Agents.Jobs.DownloadReference(jobID, resourceID, asAttachment)
client.Agents.Jobs.DeleteData(jobID)
```

## Supported Models

| Model | Value |
|-------|-------|
| GPT-5.2 | `gpt-5.2-2025-12-11` |
| GPT-5.1 | `gpt-5.1-2025-11-13` |
| GPT-5 | `gpt-5-2025-08-07` |
| GPT-5 Mini | `gpt-5-mini-2025-08-07` |
| GPT-4.1 | `gpt-4.1-2025-04-14` |
| GPT-4.1 Mini | `gpt-4.1-mini-2025-04-14` |
| O3 Pro | `o3-pro-2025-06-10` |
| O3 | `o3-2025-04-16` |
| O4 Mini | `o4-mini-2025-04-16` |
| GPT-4o | `gpt-4o-2024-11-20` |
| Grok 4 | `grok-4-0709` |
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
|--------|-----|
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
- [API Docs](https://docs.roe-ai.com)
