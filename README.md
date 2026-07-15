# Roe Go SDK

Go SDK for the [Roe](https://www.roe-ai.com/) API.

<!-- ROE-SDK:RELEASE-BANNER:START -->
> **v1.1.5** - SDK operation coverage is synchronized across Python,
> TypeScript, and Go. See `SDK_EXAMPLES.md` for copy-ready examples and use cases.
> The module path remains `github.com/roe-ai/roe-golang`.
<!-- ROE-SDK:RELEASE-BANNER:END -->

> **v1.0.0** - The Go SDK uses an OpenAPI-backed transport behind stable public
> client groups such as `Agents`, `Policies`, `Tables`, `Connections`, and
> `Connectors`.

## Installation

```bash
go get github.com/roe-ai/roe-golang@latest
```

Requires Go 1.24+

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

<!-- ROE-SDK:GENERATED-FRIENDLY-APIS:START -->
## SDK Operation Groups

Common operations are available directly on the SDK client.

```go
engines, err := client.Discovery.ListAgentEngineTypes()
models, err := client.Discovery.ListSupportedModels("text")

upload, err := client.Tables.Upload("customers", roe.FileUpload{Path: "customers.csv"}, true)
```
<!-- ROE-SDK:GENERATED-FRIENDLY-APIS:END -->

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

## Errors

Non-2xx responses return typed errors that embed `*APIError` and expose
`.StatusCode` / `.Message`. Use `errors.As` to handle expected failures
without parsing error strings:

```go
package main

import (
    "errors"
    "log"
    "os"

    roe "github.com/roe-ai/roe-golang"
)

func main() {
    client, err := roe.NewClient(
        os.Getenv("ROE_API_KEY"),
        os.Getenv("ROE_ORGANIZATION_ID"),
        "", 0, 0,
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    _, err = client.Agents.Retrieve("00000000-0000-0000-0000-000000000000")

    var notFound *roe.NotFoundError
    if errors.As(err, &notFound) {
        log.Printf("not found: %d %s", notFound.StatusCode, notFound.Message)
    }
}
```

The full hierarchy is `BadRequestError` (400), `AuthenticationError` (401),
`InsufficientCreditsError` (402), `ForbiddenError` (403), `NotFoundError`
(404), `RateLimitError` (429), and `ServerError` (5xx) — all embedding
`*APIError`.

`job.Wait(...)` does not return a typed error for agent-side failures —
instead the returned result reports `result.Failed() == true` with
`result.ErrorMessage` populated. Transport / HTTP errors hit the typed
hierarchy above.

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
            "model":       "gpt-5.5-2026-04-23",
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
    map[string]any{
        "instructions": "Investigate the alert and include layering rules.",
    },
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

// Skip the job-result cache and force a fresh run (the fresh result still
// refreshes the cache). All Run* methods accept the option.
job, _ = client.Agents.Run("agent-uuid", 0, map[string]any{"text": "input"}, nil, roe.WithSkipCache(true))

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
client.Agents.Replace(agentID, name, disableCache, cacheFailedJobs)
client.Agents.Delete(agentID)
client.Agents.Duplicate(agentID)
```

> `Agents.Duplicate(...)` returns the new `BaseAgent` directly — the new
> agent's id is on the returned value as `.ID`.
>
> Compatibility: `Agents.Update(...)` and `Agents.Versions.Update(...)` use
> PATCH for partial updates. Use `Replace(...)` when you need the PUT
> replacement endpoints.

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
client.Agents.Versions.Replace(agentID, versionID, versionName, desc)
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
| GPT-5.5 Pro | `gpt-5.5-pro-2026-04-23` |
| GPT-5.5 | `gpt-5.5-2026-04-23` |
| GPT-5.4 Pro | `gpt-5.4-pro-2026-03-05` |
| GPT-5.4 | `gpt-5.4-2026-03-05` |
| GPT-5.4 Mini | `gpt-5.4-mini-2026-03-17` |
| GPT-5.4 Nano | `gpt-5.4-nano-2026-03-17` |
| GPT-5.2 | `gpt-5.2-2025-12-11` |
| GPT-5 | `gpt-5-2025-08-07` |
| GPT-4.1 | `gpt-4.1-2025-04-14` |
| Claude Opus 4.8 | `claude-opus-4-8` |
| Claude Opus 4.7 | `claude-opus-4-7` |
| Claude Opus 4.6 | `claude-opus-4-6` |
| Claude Sonnet 4.6 | `claude-sonnet-4-6` |
| Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| Gemini 3.1 Pro | `gemini-3.1-pro-preview` |
| Gemini 3 Flash | `gemini-3-flash-preview` |
| Grok 4.20 Reasoning | `grok-4.20-0309-reasoning` |

## Engine Classes

| Engine | ID |
|--------|----|
| Multimodal Extraction | `MultimodalExtractionEngine` |
| Document Insights | `PDFExtractionEngine` |
| Document Segmentation | `PDFPageSelectionEngine` |
| Web Insights | `URLWebsiteExtractionEngine` |
| Interactive Web | `InteractiveWebExtractionEngine` |
| Web Search | `URLFinderEngine` |
| Research | `ResearchEngine` |
| Maps Search | `GoogleMapsEntityExtractionEngine` |
| Social Media | `SocialScraperEngine` |
| Marketplace Storefront Analysis | `MarketplaceStorefrontAnalysisEngine` |
| Product Compliance | `ProductPolicyEngine` |
| Merchant Risk | `MerchantRiskEngine` |
| AML Investigation | `AMLInvestigationEngine` |
| Fraud Investigation | `FraudInvestigationEngine` |

## Development

Before opening a PR, format and lint the codebase by running:

```bash
./roe-cli format
```

CI runs formatting and lint checks (`gofmt` and `golangci-lint`) on every pull request and on merges to `main`, and they must pass before a PR can be merged. Local `./roe-cli format` also runs `go vet`.

## Links

- [Roe](https://www.roe-ai.com/)
- [API Docs](https://docs.roe-ai.com)
- [Examples](examples/)
