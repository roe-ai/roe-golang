# Go SDK Examples

<!-- AUTO-GENERATED. Do not edit by hand. -->

## Examples

Copy-ready calls for every SDK operation. Required and optional inputs are shown inline in each code block.

### Agents

#### `agents_list`

List agents or create a new agent.

```go
params := &generated.AgentsListParams{
    EngineClassId: &[]string{"engine_class_id"}[0], // optional query engine_class_id
    ExcludeEngineClassId: &[]string{"exclude_engine_class_id"}[0], // optional query exclude_engine_class_id
    IncludeJobStats: &[]bool{true}[0], // optional query include_job_stats
    Ordering: &[]string{"ordering"}[0], // optional query ordering
    OrganizationId: openapi_types.UUID("00000000-0000-0000-0000-000000000000"), // required query organization_id
    Page: &[]int{1}[0], // optional query page
    PageSize: &[]int{1}[0], // optional query page_size
    Search: &[]string{"search"}[0], // optional query search
    Tags: &[]string{"value"}, // optional query tags
}

resp, err := raw.AgentsListWithResponse(
    ctx,
    params,
)
```

#### `agents_create`

Create a new base agent.

```go
params := &generated.AgentsCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: name, engine_class_id
// Optional body fields: organization_id, version_name, description, input_definitions, engine_config
body := strings.NewReader(`{
  "name": "name",
  "engine_class_id": "engine_class_id",
  "organization_id": "00000000-0000-0000-0000-000000000000",
  "version_name": "version_name",
  "description": "description",
  "input_definitions": "input_definitions",
  "engine_config": "engine_config"
}`)
contentType := "application/json"

resp, err := raw.AgentsCreateWithBodyWithResponse(
    ctx,
    params,
    contentType,
    body,
)
```

#### `agents_jobs_results_create`

Get results for multiple agent jobs

```go
params := &generated.AgentsJobsResultsCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: job_ids
body := strings.NewReader(`{
  "job_ids": ["value"]
}`)
contentType := "application/json"

resp, err := raw.AgentsJobsResultsCreateWithBodyWithResponse(
    ctx,
    params,
    contentType,
    body,
)
```

#### `agents_jobs_statuses_create`

Get status for multiple agent jobs

```go
params := &generated.AgentsJobsStatusesCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: job_ids
body := strings.NewReader(`{
  "job_ids": ["value"]
}`)
contentType := "application/json"

resp, err := raw.AgentsJobsStatusesCreateWithBodyWithResponse(
    ctx,
    params,
    contentType,
    body,
)
```

#### `agents_jobs_references_retrieve`

Serve a reference file associated with an agent job.

```go
agentJobId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_job_id
resourceId := "resource_id" // required path resource_id

params := &generated.AgentsJobsReferencesRetrieveParams{
    Download: &[]bool{true}[0], // optional query download
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsReferencesRetrieveWithResponse(
    ctx,
    agentJobId,
    resourceId,
    params,
)
```

#### `agents_jobs_result_retrieve`

Get agent job result data.

```go
agentJobId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_job_id

params := &generated.AgentsJobsResultRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsResultRetrieveWithResponse(
    ctx,
    agentJobId,
    params,
)
```

#### `agents_jobs_cancel_create`

Cancel an agent job

```go
jobId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path job_id

params := &generated.AgentsJobsCancelCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsCancelCreateWithResponse(
    ctx,
    jobId,
    params,
)
```

#### `agents_jobs_delete_data_create`

Delete agent job data

```go
jobId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path job_id

params := &generated.AgentsJobsDeleteDataCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsDeleteDataCreateWithResponse(
    ctx,
    jobId,
    params,
)
```

#### `agents_jobs_status_retrieve`

Get agent job status.

```go
jobId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path job_id

params := &generated.AgentsJobsStatusRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsStatusRetrieveWithResponse(
    ctx,
    jobId,
    params,
)
```

#### `agents_run`

Run agent synchronously

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsRunParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: metadata
body := strings.NewReader(`{
  "metadata": {}
}`)
contentType := "application/json"

resp, err := raw.AgentsRunWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_run_async_create`

Run agent asynchronously.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsRunAsyncCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: metadata
body := strings.NewReader(`{
  "metadata": {}
}`)
contentType := "application/json"

resp, err := raw.AgentsRunAsyncCreateWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_run_async_many`

Run agent asynchronously with multiple inputs

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsRunAsyncManyParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: inputs
body := strings.NewReader(`{
  "inputs": ["value"]
}`)
contentType := "application/json"

resp, err := raw.AgentsRunAsyncManyWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_run_version`

Run agent version synchronously

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsRunVersionParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: metadata
body := strings.NewReader(`{
  "metadata": {}
}`)
contentType := "application/json"

resp, err := raw.AgentsRunVersionWithBodyWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
    contentType,
    body,
)
```

#### `agents_run_versions_async_create`

Run agent version asynchronously.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsRunVersionsAsyncCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: metadata
body := strings.NewReader(`{
  "metadata": {}
}`)
contentType := "application/json"

resp, err := raw.AgentsRunVersionsAsyncCreateWithBodyWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
    contentType,
    body,
)
```

#### `agents_destroy`

Delete a base agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsDestroyParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsDestroyWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_retrieve`

Retrieve an agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsRetrieveWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_partial_update`

Partially update an agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsPartialUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: name, disable_cache, cache_failed_jobs
body := strings.NewReader(`{
  "name": "name",
  "disable_cache": true,
  "cache_failed_jobs": true
}`)
contentType := "application/json"

resp, err := raw.AgentsPartialUpdateWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_update`

Update a base agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: name, disable_cache, cache_failed_jobs
body := strings.NewReader(`{
  "name": "name",
  "disable_cache": true,
  "cache_failed_jobs": true
}`)
contentType := "application/json"

resp, err := raw.AgentsUpdateWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_duplicate_create`

Duplicate an agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsDuplicateCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsDuplicateCreateWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_jobs_cancel_all_create`

Cancel all agent jobs

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsJobsCancelAllCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsJobsCancelAllCreateWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_versions_list`

List agent versions.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsVersionsListParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsVersionsListWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_versions_create`

Create a new agent version.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsVersionsCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: version_name, description, input_definitions, engine_config
body := strings.NewReader(`{
  "version_name": "version_name",
  "description": "description",
  "input_definitions": "input_definitions",
  "engine_config": "engine_config"
}`)
contentType := "application/json"

resp, err := raw.AgentsVersionsCreateWithBodyWithResponse(
    ctx,
    agentId,
    params,
    contentType,
    body,
)
```

#### `agents_versions_current_retrieve`

Retrieve the current version of an agent.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id

params := &generated.AgentsVersionsCurrentRetrieveParams{
    GetSupportsEval: &[]bool{true}[0], // optional query get_supports_eval
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsVersionsCurrentRetrieveWithResponse(
    ctx,
    agentId,
    params,
)
```

#### `agents_versions_destroy`

Delete an agent version.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsVersionsDestroyParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsVersionsDestroyWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
)
```

#### `agents_versions_retrieve`

Retrieve an agent version.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsVersionsRetrieveParams{
    GetSupportsEval: &[]bool{true}[0], // optional query get_supports_eval
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.AgentsVersionsRetrieveWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
)
```

#### `agents_versions_partial_update`

Partially update an agent version.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsVersionsPartialUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: version_name, description
body := strings.NewReader(`{
  "version_name": "version_name",
  "description": "description"
}`)
contentType := "application/json"

resp, err := raw.AgentsVersionsPartialUpdateWithBodyWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
    contentType,
    body,
)
```

#### `agents_versions_update`

Update an agent version.

```go
agentId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_id
agentVersionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path agent_version_id

params := &generated.AgentsVersionsUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: version_name, description
body := strings.NewReader(`{
  "version_name": "version_name",
  "description": "description"
}`)
contentType := "application/json"

resp, err := raw.AgentsVersionsUpdateWithBodyWithResponse(
    ctx,
    agentId,
    agentVersionId,
    params,
    contentType,
    body,
)
```

### Connections

#### `connections_list`

List/create connections.

```go
params := &generated.ConnectionsListParams{
    ConnectorType: &[]string{"connector_type"}[0], // optional query connector_type
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
    Page: &[]int{1}[0], // optional query page
    PageSize: &[]int{1}[0], // optional query page_size
    Search: &[]string{"search"}[0], // optional query search
}

resp, err := raw.ConnectionsListWithResponse(
    ctx,
    params,
)
```

#### `connections_create`

List/create connections.

```go
params := &generated.ConnectionsCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: connector_type, name, config
// Optional body fields: description, auth_config, organization_id
body := strings.NewReader(`{
  "connector_type": "connector_type",
  "name": "name",
  "description": "description",
  "config": {},
  "auth_config": {},
  "organization_id": "00000000-0000-0000-0000-000000000000"
}`)
contentType := "application/json"

resp, err := raw.ConnectionsCreateWithBodyWithResponse(
    ctx,
    params,
    contentType,
    body,
)
```

#### `connections_test_credentials_create`

Test credentials without storing them.

```go
// Required body fields: connector_type, config
// Optional body fields: auth_config
body := strings.NewReader(`{
  "connector_type": "connector_type",
  "config": {},
  "auth_config": {}
}`)
contentType := "application/json"

resp, err := raw.ConnectionsTestCredentialsCreateWithBodyWithResponse(
    ctx,
    contentType,
    body,
)
```

#### `connections_destroy`

Manage connection.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.ConnectionsDestroyParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.ConnectionsDestroyWithResponse(
    ctx,
    id,
    params,
)
```

#### `connections_retrieve`

Manage connection.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.ConnectionsRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.ConnectionsRetrieveWithResponse(
    ctx,
    id,
    params,
)
```

#### `connections_partial_update`

Manage connection.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.ConnectionsPartialUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: name, description, config, auth_config
body := strings.NewReader(`{
  "name": "name",
  "description": "description",
  "config": {},
  "auth_config": {}
}`)
contentType := "application/json"

resp, err := raw.ConnectionsPartialUpdateWithBodyWithResponse(
    ctx,
    id,
    params,
    contentType,
    body,
)
```

#### `connections_update`

Manage connection.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.ConnectionsUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: name, description, config, auth_config
body := strings.NewReader(`{
  "name": "name",
  "description": "description",
  "config": {},
  "auth_config": {}
}`)
contentType := "application/json"

resp, err := raw.ConnectionsUpdateWithBodyWithResponse(
    ctx,
    id,
    params,
    contentType,
    body,
)
```

#### `connections_test_create`

Test connection.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.ConnectionsTestCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.ConnectionsTestCreateWithResponse(
    ctx,
    id,
    params,
)
```

### Connectors

#### `connectors_retrieve`

List all connector types.

```go
resp, err := raw.ConnectorsRetrieveWithResponse(
    ctx,
)
```

#### `connectors_retrieve_by_type`

Get connector details.

```go
connectorType := "connector_type" // required path connector_type

resp, err := raw.ConnectorsRetrieveByTypeWithResponse(
    ctx,
    connectorType,
)
```

### Discovery

#### `discovery_supported_models_list`

List supported model IDs

```go
params := &generated.DiscoverySupportedModelsListParams{
    Capability: &[]string{"capability"}[0], // optional query capability
}

resp, err := raw.DiscoverySupportedModelsListWithResponse(
    ctx,
    params,
)
```

#### `discovery_agent_engine_types_list`

List supported agent engine types

```go
resp, err := raw.DiscoveryAgentEngineTypesListWithResponse(
    ctx,
)
```

### Policies

#### `policies_list`

List all policies and create a new policy.

```go
params := &generated.PoliciesListParams{
    Ordering: &[]string{"ordering"}[0], // optional query ordering
    Page: &[]int{1}[0], // optional query page
    PageSize: &[]int{1}[0], // optional query page_size
    Search: &[]string{"search"}[0], // optional query search
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.PoliciesListWithResponse(
    ctx,
    params,
)
```

#### `policies_create`

List all policies and create a new policy.

```go
params := &generated.PoliciesCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: name, content
// Optional body fields: description, version_name
body := strings.NewReader(`{
  "name": "name",
  "description": "description",
  "content": "content",
  "version_name": "version_name"
}`)
contentType := "application/json"

resp, err := raw.PoliciesCreateWithBodyWithResponse(
    ctx,
    params,
    contentType,
    body,
)
```

#### `policies_destroy`

Retrieve, update, or delete a single policy by ID.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.PoliciesDestroyParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.PoliciesDestroyWithResponse(
    ctx,
    id,
    params,
)
```

#### `policies_retrieve`

Retrieve, update, or delete a single policy by ID.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.PoliciesRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.PoliciesRetrieveWithResponse(
    ctx,
    id,
    params,
)
```

#### `policies_partial_update`

Retrieve, update, or delete a single policy by ID.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.PoliciesPartialUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Optional body fields: name, description
body := strings.NewReader(`{
  "name": "name",
  "description": "description"
}`)
contentType := "application/json"

resp, err := raw.PoliciesPartialUpdateWithBodyWithResponse(
    ctx,
    id,
    params,
    contentType,
    body,
)
```

#### `policies_update`

Retrieve, update, or delete a single policy by ID.

```go
id := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path id

params := &generated.PoliciesUpdateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: name
// Optional body fields: description
body := strings.NewReader(`{
  "name": "name",
  "description": "description"
}`)
contentType := "application/json"

resp, err := raw.PoliciesUpdateWithBodyWithResponse(
    ctx,
    id,
    params,
    contentType,
    body,
)
```

#### `policies_versions_list`

Create a new policy version or list all versions of a specific policy.

```go
policyId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path policy_id

params := &generated.PoliciesVersionsListParams{
    Page: &[]int{1}[0], // optional query page
    PageSize: &[]int{1}[0], // optional query page_size
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.PoliciesVersionsListWithResponse(
    ctx,
    policyId,
    params,
)
```

#### `policies_versions_create`

Create a new policy version or list all versions of a specific policy.

```go
policyId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path policy_id

params := &generated.PoliciesVersionsCreateParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}
// Required body fields: content
// Optional body fields: version_name, base_version_id
body := strings.NewReader(`{
  "version_name": "version_name",
  "content": "content",
  "base_version_id": "00000000-0000-0000-0000-000000000000"
}`)
contentType := "application/json"

resp, err := raw.PoliciesVersionsCreateWithBodyWithResponse(
    ctx,
    policyId,
    params,
    contentType,
    body,
)
```

#### `policies_versions_retrieve`

Get a specific policy version by policy_id and version_id.

```go
policyId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path policy_id
versionId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path version_id

params := &generated.PoliciesVersionsRetrieveParams{
    OrganizationId: &[]openapi_types.UUID{openapi_types.UUID("00000000-0000-0000-0000-000000000000")}[0], // optional query organization_id
}

resp, err := raw.PoliciesVersionsRetrieveWithResponse(
    ctx,
    policyId,
    versionId,
    params,
)
```

### Tables

#### `tables_list`

List Roe tables

```go
resp, err := raw.TablesListWithResponse(
    ctx,
)
```

#### `tables_query_create`

Run a read-only Roe table query

```go
// Required body fields: sql
// Optional body fields: limit
body := strings.NewReader(`{
  "sql": "sql",
  "limit": 1
}`)
contentType := "application/json"

resp, err := raw.TablesQueryCreateWithBodyWithResponse(
    ctx,
    contentType,
    body,
)
```

#### `tables_query_result_retrieve`

Get a Roe table query result

```go
tableQueryId := openapi_types.UUID("00000000-0000-0000-0000-000000000000") // required path table_query_id

resp, err := raw.TablesQueryResultRetrieveWithResponse(
    ctx,
    tableQueryId,
)
```

#### `tables_destroy`

Delete a Roe table

```go
tableName := "table_name" // required path table_name

resp, err := raw.TablesDestroyWithResponse(
    ctx,
    tableName,
)
```

#### `tables_describe_retrieve`

Describe a Roe table

```go
tableName := "table_name" // required path table_name

resp, err := raw.TablesDescribeRetrieveWithResponse(
    ctx,
    tableName,
)
```

#### `tables_preview_retrieve`

Preview a Roe table

```go
tableName := "table_name" // required path table_name

params := &generated.TablesPreviewRetrieveParams{
    Limit: &[]int{1}[0], // optional query limit
}

resp, err := raw.TablesPreviewRetrieveWithResponse(
    ctx,
    tableName,
    params,
)
```

#### `upload_table`

Upload a CSV as a Roe table

```go
var body bytes.Buffer
writer := multipart.NewWriter(&body)
// required body field table_name
if err := writer.WriteField("table_name", "table_name"); err != nil {
    panic(err)
}
// required body field file
fileReader, err := os.Open("file.csv")
if err != nil {
    panic(err)
}
defer fileReader.Close()
fileWriter, err := writer.CreateFormFile("file", "file.csv")
if err != nil {
    panic(err)
}
if _, err = io.Copy(fileWriter, fileReader); err != nil {
    panic(err)
}
// optional body field with_headers
if err := writer.WriteField("with_headers", "true"); err != nil {
    panic(err)
}
// optional body field organization_id
if err := writer.WriteField("organization_id", "00000000-0000-0000-0000-000000000000"); err != nil {
    panic(err)
}
if err := writer.Close(); err != nil {
    panic(err)
}
contentType := writer.FormDataContentType()

resp, err := raw.UploadTableWithBodyWithResponse(
    ctx,
    contentType,
    &body,
)
```

### Users

#### `users_current_user_retrieve`

Get the current user

```go
resp, err := raw.UsersCurrentUserRetrieveWithResponse(
    ctx,
)
```

## Use Cases

These workflows assume `ROE_API_KEY` and `ROE_ORGANIZATION_ID` are set.

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
