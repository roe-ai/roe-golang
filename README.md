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
    }, nil)
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

## Job Result Inspection

After waiting for a job, you can inspect its outcome using status helpers:

```go
result, err := job.Wait(5*time.Second, 0)
if err != nil {
    log.Fatal(err) // only for timeouts/network errors
}

if result.Succeeded() {
    for _, output := range result.Outputs {
        fmt.Printf("%s: %s\n", output.Key, output.Value)
    }
} else if result.Cancelled() {
    fmt.Println("Job was cancelled")
} else if result.Failed() {
    if result.ErrorMessage != nil {
        fmt.Println("Error:", *result.ErrorMessage)
    }
}

// Available fields
// result.Status        *JobStatus - set by Wait/WaitContext
// result.ErrorMessage  *string    - error details if failed
// result.Outputs       []AgentDatum
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
    }, nil)
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

## Rori Agents (Agentic Workflows)

Rori agents are autonomous investigation agents that follow policies (SOPs), use tools, and produce structured verdicts. Unlike extraction engines which transform data, Rori agents reason over evidence, apply policy rules, and return dispositions.

### Policies

Policies define the rules, instructions, and disposition classifications that Rori agents follow:

```go
// Create a policy with an initial version
policy, _ := client.Policies.Create(
    "AML Investigation Policy",
    map[string]any{
        "guidelines": map[string]any{
            "categories": []any{
                map[string]any{
                    "title": "Structuring",
                    "rules": []any{
                        map[string]any{
                            "title":       "Cash structuring below reporting thresholds",
                            "description": "Multiple deposits just under $10,000",
                            "flag":        "RED_FLAG",
                        },
                    },
                },
            },
        },
        "instructions": "Investigate the alert against each category.",
        "dispositions": map[string]any{
            "classifications": []any{
                map[string]any{"name": "Suspicious", "description": "Activity warrants SAR filing"},
                map[string]any{"name": "Not Suspicious", "description": "Legitimate activity"},
                map[string]any{"name": "Needs Escalation", "description": "Requires senior review"},
            },
        },
    },
    "Standard operating procedure for AML investigation",
    "v1",
)
```

Iterate on policies by creating new versions:

```go
// Create a new version (automatically becomes the current version)
newVersion, _ := client.Policies.Versions.Create(
    policy.ID,
    map[string]any{...}, // Updated policy content
    "v2 - added layering rules",
    "",
)

// List all versions
versions, _ := client.Policies.Versions.List(policy.ID)

// Retrieve a specific version
version, _ := client.Policies.Versions.Retrieve(policy.ID, newVersion.ID)

// Update policy metadata
name := "Updated Policy Name"
client.Policies.Update(policy.ID, &name, nil)

// List all policies
policies, _ := client.Policies.List(1, 50)

// Delete a policy
client.Policies.Delete(policy.ID)
```

### Policy Content Reference

| Field | Type | Description |
|-------|------|-------------|
| `guidelines` | object | Categories -> Rules -> Sub-rules hierarchy |
| `guidelines.categories[].title` | string | Category name |
| `guidelines.categories[].rules[].title` | string | Rule name |
| `guidelines.categories[].rules[].description` | string | Rule details |
| `guidelines.categories[].rules[].flag` | string | `"RED_FLAG"` or `"GREEN_FLAG"` |
| `instructions` | string | Free-text investigation instructions |
| `dispositions.classifications[].name` | string | Outcome label (e.g., "Suspicious") |
| `dispositions.classifications[].description` | string | When to apply this outcome |
| `summary_template.template` | string | Handlebars template for report generation |

### Rori Agent Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `policy_version_id` | string | -- | Policy version UUID (required) |
| `context_sources` | list | `[]` | External data sources (SQL connections, APIs) |
| `enable_planning` | bool | `true` | Enable autonomous tool-use planning |
| `enable_memory` | bool | `false` | Retain context across runs for the same entity |
| `reasoning_effort` | string | `"medium"` | `"low"`, `"medium"`, or `"high"` |

## Running Agents

```go
// Async (recommended)
job, _ := client.Agents.Run("agent-uuid", 0, map[string]any{"text": "input"}, nil)
result, _ := job.Wait(5*time.Second, 0)

// Sync
outputs, _ := client.Agents.RunSync("agent-uuid", map[string]any{"text": "input"}, nil)

// With files (auto-uploaded)
job, _ := client.Agents.Run("agent-uuid", 0, map[string]any{
    "document": "/path/to/file.pdf",
}, nil)

// Batch processing
batch, _ := client.Agents.RunMany("agent-uuid", []map[string]any{
    {"text": "input1"},
    {"text": "input2"},
}, 0, nil)
results, _ := batch.Wait(5*time.Second, 0)

// Run a specific version
job, _ := client.Agents.RunVersion("agent-uuid", "version-uuid", 0, map[string]any{
    "text": "input",
}, nil)
```

## Metadata

Attach arbitrary metadata to any job when running an agent. Metadata is stored with the job for tracking and correlation.

```go
// Attach metadata to an async job
job, _ := client.Agents.Run("agent-uuid", 0, map[string]any{
    "url": "https://example.com",
}, map[string]any{
    "customer_id":    "cust-123",
    "request_source": "api",
})

// Attach metadata to a batch of jobs
batch, _ := client.Agents.RunMany("agent-uuid", []map[string]any{
    {"url": "https://a.com"},
    {"url": "https://b.com"},
}, 0, map[string]any{
    "campaign": "weekly-scan",
})
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
client.Agents.Run(agentID, timeout, inputs, metadata)
client.Agents.RunSync(agentID, inputs, metadata)
client.Agents.RunMany(agentID, batchInputs, timeout, metadata)
client.Agents.RunVersion(agentID, versionID, timeout, inputs, metadata)
client.Agents.RunVersionSync(agentID, versionID, inputs, metadata)
```

### Versions

```go
client.Agents.Versions.List(agentID)
client.Agents.Versions.Retrieve(agentID, versionID, getSupportsEval)
client.Agents.Versions.RetrieveCurrent(agentID)
client.Agents.Versions.Create(agentID, inputDefs, engineConfig, versionName, desc)
client.Agents.Versions.Update(agentID, versionID, versionName, desc)
client.Agents.Versions.Delete(agentID, versionID)
```

### Jobs

```go
client.Agents.Jobs.RetrieveStatus(jobID)
client.Agents.Jobs.RetrieveResult(jobID)
client.Agents.Jobs.RetrieveStatusMany(jobIDs)
client.Agents.Jobs.RetrieveResultMany(jobIDs)
client.Agents.Jobs.DownloadReference(jobID, resourceID, asAttachment)
client.Agents.Jobs.DeleteData(jobID)
client.Agents.Jobs.Cancel(jobID)
client.Agents.Jobs.CancelAll(agentID)
```

### Policies

```go
client.Policies.List(page, pageSize)
client.Policies.Retrieve(policyID)
client.Policies.Create(name, content, description, versionName)
client.Policies.Update(policyID, name, description)
client.Policies.Delete(policyID)
```

### Policy Versions

```go
client.Policies.Versions.List(policyID)
client.Policies.Versions.Retrieve(policyID, versionID)
client.Policies.Versions.Create(policyID, content, versionName, baseVersionID)
```

## Supported Models

| Model | Value |
|-------|-------|
| GPT-5.4 | `gpt-5.4-2026-03-05` |
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
| Grok 4.1 Fast Reasoning | `grok-4-1-fast-reasoning` |
| Claude Opus 4.6 | `claude-opus-4-6` |
| Claude Sonnet 4.6 | `claude-sonnet-4-6` |
| Claude Opus 4.5 | `claude-opus-4-5-20251101` |
| Claude Sonnet 4.5 | `claude-sonnet-4-5-20250929` |
| Claude Opus 4.1 | `claude-opus-4-1-20250805` |
| Claude Opus 4 | `claude-opus-4-20250514` |
| Claude Sonnet 4 | `claude-sonnet-4-20250514` |
| Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| Claude 3.5 Haiku | `claude-3-5-haiku-20241022` |
| Gemini 3 Pro | `gemini-3-pro-preview` |
| Gemini 3 Flash | `gemini-3-flash-preview` |
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
| LinkedIn Crawler | `LinkedInScraperEngine` |
| Social Media | `SocialScraperEngine` |
| Product Compliance | `ProductPolicyEngine` |
| Merchant Risk | `MerchantRiskEngine` |
| AML Investigation | `AMLInvestigationEngine` |
| Fraud Investigation | `FraudInvestigationEngine` |

## Links

- [Roe AI](https://www.roe-ai.com/)
- [API Docs](https://docs.roe-ai.com)
- [Examples](examples/)
