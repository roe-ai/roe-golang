# Go SDK Examples

<!-- AUTO-GENERATED. Do not edit by hand. -->

## Examples

Copy-ready calls for every SDK operation. Required and optional inputs are shown inline in each code block.

### Agents

#### `agents_list`

List agents or create a new agent.

```go
result, err := client.Agents.List(
    1,
    1,
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_create`

Create a new base agent.

```go
result, err := client.Agents.Create(
    "name",
    "engineClassID",
    []map[string]any{{"key": "text", "data_type": "text/plain"}},
    map[string]any{"model": "gpt-4.1"},
    "versionName",
    "description",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_results_create`

Get results for multiple agent jobs

```go
result, err := client.Agents.Jobs.RetrieveResultMany([]string{"job_id"})
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_statuses_create`

Get status for multiple agent jobs

```go
result, err := client.Agents.Jobs.RetrieveStatusMany([]string{"job_id"})
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_references_retrieve`

Serve a reference file associated with an agent job.

```go
content, err := client.Agents.Jobs.DownloadReference(
    "job_id",
    "resource_id",
    false,
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_result_retrieve`

Get agent job result data.

```go
result, err := client.Agents.Jobs.RetrieveResult(
    "jobID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_cancel_create`

Cancel an agent job

```go
err := client.Agents.Jobs.Cancel(
    "jobID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_delete_data_create`

Delete agent job data

```go
result, err := client.Agents.Jobs.DeleteData(
    "jobID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_status_retrieve`

Get agent job status.

```go
result, err := client.Agents.Jobs.RetrieveStatus(
    "jobID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_run`

Run agent synchronously

```go
result, err := client.Agents.RunSync(
    "agent_id",
    map[string]any{"text": "text"},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_run_async_create`

Run agent asynchronously.

```go
job, err := client.Agents.Run(
    "agent_id",
    300,
    map[string]any{"text": "text"},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_run_async_many`

Run agent asynchronously with multiple inputs

```go
batch, err := client.Agents.RunMany(
    "agent_id",
    []map[string]any{{"text": "text"}},
    300,
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_run_version`

Run agent version synchronously

```go
result, err := client.Agents.RunVersionSync(
    "agent_id",
    "version_id",
    map[string]any{"text": "text"},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_run_versions_async_create`

Run agent version asynchronously.

```go
job, err := client.Agents.RunVersion(
    "agent_id",
    "version_id",
    300,
    map[string]any{"text": "text"},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_destroy`

Delete a base agent.

```go
err := client.Agents.Delete(
    "agentID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_retrieve`

Retrieve an agent.

```go
result, err := client.Agents.Retrieve(
    "agentID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_partial_update`

Partially update an agent.

```go
result, err := client.Agents.Update(
    "agentID",
    "name",
    &[]bool{true}[0],
    &[]bool{true}[0],
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_update`

Update a base agent.

```go
result, err := client.Agents.Replace(
    "agentID",
    "name",
    &[]bool{true}[0],
    &[]bool{true}[0],
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_duplicate_create`

Duplicate an agent.

```go
result, err := client.Agents.Duplicate("agent_id")
if err != nil {
    log.Fatal(err)
}
```

#### `agents_jobs_cancel_all_create`

Cancel all agent jobs

```go
err := client.Agents.Jobs.CancelAll(
    "agentID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_list`

List agent versions.

```go
params := &roe.ListVersionsParams{Page: 1, PageSize: 10}

result, err := client.Agents.Versions.ListPaginated(
    "agent_id",
    params,
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_create`

Create a new agent version.

```go
result, err := client.Agents.Versions.Create(
    "agentID",
    []map[string]any{{"key": "text", "data_type": "text/plain"}},
    map[string]any{"model": "gpt-4.1"},
    "versionName",
    "description",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_current_retrieve`

Retrieve the current version of an agent.

```go
getSupportsEval := true

result, err := client.Agents.Versions.RetrieveCurrentWithEval(
    "agent_id",
    &getSupportsEval,
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_destroy`

Delete an agent version.

```go
err := client.Agents.Versions.Delete(
    "agentID",
    "versionID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_retrieve`

Retrieve an agent version.

```go
result, err := client.Agents.Versions.Retrieve(
    "agentID",
    "versionID",
    &[]bool{true}[0],
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_partial_update`

Partially update an agent version.

```go
err := client.Agents.Versions.Update(
    "agentID",
    "versionID",
    "versionName",
    "description",
)
if err != nil {
    log.Fatal(err)
}
```

#### `agents_versions_update`

Update an agent version.

```go
result, err := client.Agents.Versions.Replace(
    "agentID",
    "versionID",
    "versionName",
    "description",
)
if err != nil {
    log.Fatal(err)
}
```

### Connections

#### `connections_list`

List/create connections.

```go
result, err := client.Connections.List(
    "connectorType",
    "search",
    1,
    1,
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_create`

List/create connections.

```go
result, err := client.Connections.Create(
    "connectorType",
    "name",
    map[string]any{},
    "description",
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_test_credentials_create`

Test credentials without storing them.

```go
result, err := client.Connections.TestCredentials(
    "connectorType",
    map[string]any{},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_destroy`

Manage connection.

```go
err := client.Connections.Delete(
    "connectionID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_retrieve`

Manage connection.

```go
result, err := client.Connections.Retrieve(
    "connectionID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_partial_update`

Manage connection.

```go
result, err := client.Connections.Update(
    "connectionID",
    "name",
    "description",
    map[string]any{},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_update`

Manage connection.

```go
result, err := client.Connections.Replace(
    "connectionID",
    "name",
    "description",
    map[string]any{},
    map[string]any{},
)
if err != nil {
    log.Fatal(err)
}
```

#### `connections_test_create`

Test connection.

```go
result, err := client.Connections.Test(
    "connectionID",
)
if err != nil {
    log.Fatal(err)
}
```

### Connectors

#### `connectors_retrieve`

List all connector types.

```go
result, err := client.Connectors.List(
)
if err != nil {
    log.Fatal(err)
}
```

#### `connectors_retrieve_by_type`

Get connector details.

```go
result, err := client.Connectors.Retrieve(
    "connectorType",
)
if err != nil {
    log.Fatal(err)
}
```

### Discovery

#### `discovery_supported_models_list`

List supported model IDs

```go
result, err := client.Discovery.ListSupportedModels(
    "capability",
)
if err != nil {
    log.Fatal(err)
}
```

#### `discovery_agent_engine_types_list`

List supported agent engine types

```go
result, err := client.Discovery.ListAgentEngineTypes(
)
if err != nil {
    log.Fatal(err)
}
```

### Policies

#### `policies_list`

List all policies and create a new policy.

```go
result, err := client.Policies.List(
    1,
    1,
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_create`

List all policies and create a new policy.

```go
result, err := client.Policies.Create(
    "name",
    map[string]any{},
    "description",
    "versionName",
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_destroy`

Retrieve, update, or delete a single policy by ID.

```go
err := client.Policies.Delete(
    "policyID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_retrieve`

Retrieve, update, or delete a single policy by ID.

```go
result, err := client.Policies.Retrieve(
    "policyID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_partial_update`

Retrieve, update, or delete a single policy by ID.

```go
result, err := client.Policies.Update(
    "policyID",
    &[]string{"value"}[0],
    &[]string{"value"}[0],
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_update`

Retrieve, update, or delete a single policy by ID.

```go
result, err := client.Policies.Replace(
    "policyID",
    "name",
    "description",
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_versions_list`

Create a new policy version or list all versions of a specific policy.

```go
result, err := client.Policies.Versions.List("policy_id")
if err != nil {
    log.Fatal(err)
}
```

#### `policies_versions_create`

Create a new policy version or list all versions of a specific policy.

```go
result, err := client.Policies.Versions.Create(
    "policy_id",
    map[string]any{},
    "version_name",
    "base_version_id",
)
if err != nil {
    log.Fatal(err)
}
```

#### `policies_versions_retrieve`

Get a specific policy version by policy_id and version_id.

```go
result, err := client.Policies.Versions.Retrieve(
    "policy_id",
    "version_id",
)
if err != nil {
    log.Fatal(err)
}
```

### Tables

#### `tables_list`

List Roe tables

```go
result, err := client.Tables.List(
)
if err != nil {
    log.Fatal(err)
}
```

#### `tables_query_create`

Run a read-only Roe table query

```go
result, err := client.Tables.Query(
    "sql",
    1,
)
if err != nil {
    log.Fatal(err)
}
```

#### `tables_query_result_retrieve`

Get a Roe table query result

```go
result, err := client.Tables.QueryResult(
    "tableQueryID",
)
if err != nil {
    log.Fatal(err)
}
```

#### `tables_destroy`

Delete a Roe table

```go
err := client.Tables.Delete(
    "tableName",
)
if err != nil {
    log.Fatal(err)
}
```

#### `tables_describe_retrieve`

Describe a Roe table

```go
result, err := client.Tables.Describe(
    "tableName",
)
if err != nil {
    log.Fatal(err)
}
```

#### `tables_preview_retrieve`

Preview a Roe table

```go
result, err := client.Tables.Preview(
    "tableName",
    1,
)
if err != nil {
    log.Fatal(err)
}
```

#### `upload_table`

Upload a CSV as a Roe table

```go
file := roe.FileUpload{Path: "file.csv"}

result, err := client.Tables.Upload(
    "table_name",
    file,
    true,
)
if err != nil {
    log.Fatal(err)
}
```

### Users

#### `users_current_user_retrieve`

Get the current user

```go
result, err := client.Users.Me()
if err != nil {
    log.Fatal(err)
}
```

## Use Cases

These workflows assume `ROE_API_KEY` and `ROE_ORGANIZATION_ID` are set.
The first example provisions a policy and an agent from scratch; the later two
reuse an existing agent id so they stay focused on the run-and-fetch calls.

### Create a policy and run a policy-aware agent

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
    client, err := roe.NewClient(os.Getenv("ROE_API_KEY"), os.Getenv("ROE_ORGANIZATION_ID"), "", 0, 0)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    policy, err := client.Policies.Create("AML Investigation Policy", map[string]any{
        "guidelines": map[string]any{
            "categories": []map[string]any{
                {
                    "title": "Transaction Patterns",
                    "rules": []map[string]any{
                        {
                            "title":       "Structuring below reporting thresholds",
                            "flag":        "RED_FLAG",
                            "description": "Deposits just under CTR thresholds in a short window.",
                        },
                    },
                },
            },
        },
        "dispositions": map[string]any{
            "classifications": []map[string]any{
                {"name": "SAR", "description": "File a Suspicious Activity Report."},
                {"name": "DISMISS", "description": "Close as non-suspicious."},
            },
        },
    }, "", "")
    if err != nil {
        log.Fatal(err)
    }

    agent, err := client.Agents.Create(
        "AML Investigation Agent",
        "AMLInvestigationEngine",
        []map[string]any{{"key": "alert_data", "data_type": "text/plain", "description": "Alert to investigate."}},
        map[string]any{"policy_version_id": *policy.CurrentVersionID, "alert_data": "${alert_data}"},
        "", "",
    )
    if err != nil {
        log.Fatal(err)
    }

    job, err := client.Agents.Run(agent.ID, 300, map[string]any{
        "alert_data": "Customer made 9 cash deposits of $9,500 over three days.",
    }, nil)
    if err != nil {
        log.Fatal(err)
    }

    result, err := job.Wait(5*time.Second, 5*time.Minute)
    if err != nil {
        log.Fatal(err)
    }

    for _, output := range result.Outputs {
        fmt.Printf("%s: %s\n", output.Key, output.Value)
    }
}
```

### Run an agent and download a saved reference

```go
package main

import (
    "encoding/json"
    "log"
    "os"
    "time"

    roe "github.com/roe-ai/roe-golang"
)

type reference struct {
    ResourceID string `json:"resource_id"`
}

func main() {
    client, err := roe.NewClient(os.Getenv("ROE_API_KEY"), os.Getenv("ROE_ORGANIZATION_ID"), "", 0, 0)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    agentID := os.Getenv("ROE_URL_AGENT_ID")
    if agentID == "" {
        log.Fatal("Set ROE_URL_AGENT_ID")
    }

    job, err := client.Agents.Run(agentID, 300, map[string]any{"url": "https://www.roe-ai.com/"}, map[string]any{"use_case": "website-scan"})
    if err != nil {
        log.Fatal(err)
    }

    result, err := job.Wait(5*time.Second, 5*time.Minute)
    if err != nil {
        log.Fatal(err)
    }

    for _, output := range result.Outputs {
        for _, ref := range referencesFrom(output.Value) {
            content, err := client.Agents.Jobs.DownloadReference(job.ID(), ref.ResourceID, false)
            if err != nil {
                log.Fatal(err)
            }
            if err := os.WriteFile(ref.ResourceID+".bin", content, 0644); err != nil {
                log.Fatal(err)
            }
        }
    }
}

func referencesFrom(value string) []reference {
    var payload struct {
        References []reference `json:"references"`
    }
    if err := json.Unmarshal([]byte(value), &payload); err != nil {
        return nil
    }
    return payload.References
}
```

### Run a batch of inputs

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
    client, err := roe.NewClient(os.Getenv("ROE_API_KEY"), os.Getenv("ROE_ORGANIZATION_ID"), "", 0, 0)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    agentID := os.Getenv("ROE_TEXT_AGENT_ID")
    if agentID == "" {
        log.Fatal("Set ROE_TEXT_AGENT_ID")
    }

    batch, err := client.Agents.RunMany(agentID, []map[string]any{
        {"text": "Summarize the customer complaint."},
        {"text": "Extract the requested follow-up action."},
    }, 300, nil)
    if err != nil {
        log.Fatal(err)
    }

    results, err := batch.Wait(5*time.Second, 5*time.Minute)
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        for _, output := range result.Outputs {
            fmt.Printf("%s: %s\n", output.Key, output.Value)
        }
    }
}
```
