# Changelog

All notable changes to the Roe AI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-04-28

The SDK's hand-written `agents.go` / `policies.go` / `http.go` request layer
has been replaced by a thin facade over the generated OpenAPI client (oapi-
codegen, in `github.com/roe-ai/roe-golang/generated`). Public method names
and signatures on `AgentsAPI`, `PoliciesAPI`, and their nested APIs are
preserved; the **return types** now come from the `generated` package.

### Added

- `dynamicInputsRequest(inputs, metadata) -> (io.Reader, contentType, error)`
  on `httpClient` — pure pre-processor, fed into the generated client's
  `*WithBodyWithResponse` helpers for multipart agent runs.
- `retryDoer` adapter implementing `generated.HttpRequestDoer`, wired through
  `RoeClient.Raw()` so direct callers of the generated client now benefit
  from the same exponential-backoff retry policy and `errorFromResponse`
  translation as the wrapper layer.
- `errorFromResponse(*http.Response, []byte) error` helper in `errors.go`
  for translating generated responses into the existing typed error tree
  (`BadRequestError`, `RateLimitError`, etc.).

### Changed (breaking)

- **Return types** on every method in `AgentsAPI`, `AgentVersionsAPI`,
  `AgentJobsAPI`, `PoliciesAPI`, `PolicyVersionsAPI` are now generated types
  (`*generated.PaginatedBaseAgentList`, `*generated.BaseAgent`, `*generated.
  Policy`, etc.) instead of the hand-written `BaseAgent` / `Policy` /
  `PaginatedResponse[T]` structs. Field names follow oapi-codegen
  conventions (e.g. `Id` rather than `ID`, pointer types for optional
  fields).
- `Update` on both `AgentsAPI` and `PoliciesAPI` (and the nested Versions
  APIs) now issues `PATCH` instead of `PUT`. The generated `PUT` body
  requires every field; `PATCH` matches the existing partial-update
  ergonomics. Wire-level breaking change for downstream callers that
  inspected the request method.
- `Run`, `RunSync`, `RunVersion`, `RunVersionSync` now build the multipart
  body via `dynamicInputsRequest` and submit through
  `V1AgentsRunAsyncCreateWithBodyWithResponse`. The previous code path went
  through `httpClient.postDynamicInputs`, which has been deleted.
- `RunMany`, `RetrieveStatusMany`, `RetrieveResultMany` deliberately bypass
  the typed `*WithResponse` parser and use the lower-level `*WithBody`
  variant because the OpenAPI spec disagrees with the production wire
  format on these three endpoints (run-many: spec=`object`/wire=`[]string`;
  bulk results: spec lacks per-item ids; bulk statuses: same).
- `PolicyVersionsAPI.List` now returns the generated paginated wrapper
  (`*generated.PaginatedPolicyVersionList`) instead of a flat `[]PolicyVersion`.

### Removed (breaking)

- `httpClient.{get, getWithContext, getBytesWithContext, deleteWithContext,
  postJSONWithContext, putJSONWithContext, postDynamicInputs,
  postDynamicInputsWithContext}` — replaced by `doRequest`/`doRetried` plus
  `dynamicInputsRequest` and the generated client's typed methods.
- Hand-written response models in `types.go` (`Policy`, `PolicyVersion`,
  `PaginatedResponse[T]` and its helpers). The names `BaseAgent`,
  `AgentVersion`, `AgentInputDefinition`, `UserInfo`, `Policy`,
  `PolicyVersion` continue to resolve at the package root via type aliases
  to the equivalent `generated.*` types — call sites that only used those
  identifiers as types keep working.
- `AgentVersionsAPI.ListPaginated` / `ListPaginatedWithContext` and the
  `ListVersionsParams` struct (re-exported from `roe/reexport.go`).
  `AgentVersionsAPI.List` already returns the generated paginated wrapper
  (`*generated.PaginatedAgentVersionList`) with `Count`/`Next`/`Previous`/
  `Results` fields, so the separate paginated entry point is redundant.
  Callers that need page-level control can call the generated client
  directly via `client.Raw().V1AgentsVersionsListWithResponse(ctx, ...)`.

### Migration guide

Most callers need no change beyond pulling in v2 and updating field-access
patterns where struct fields were renamed by the codegen (e.g. `agent.ID`
becomes `agent.Id`, `*string` instead of `string` for nullable fields).

```go
// Before (v1)
agent, err := client.Agents.Retrieve(agentID)
fmt.Println(agent.ID, agent.Name)

// After (v2)
agent, err := client.Agents.Retrieve(agentID)         // returns *generated.BaseAgent
fmt.Println(*agent.Id, *agent.Name)
```

```go
// Before (v1) — hand-written list type
var resp PaginatedResponse[BaseAgent] = ...

// After (v2) — generated type
var resp *generated.PaginatedBaseAgentList = ...
// roe.BaseAgent still works (it's an alias for generated.BaseAgent).
```

Direct callers of the generated client via `client.Raw()` are unchanged in
shape but now opt into the SDK's retry policy and typed-error translation
automatically.

### Known follow-ups

- `e2e_test.go` is gated behind the `e2e_legacy` build tag for this
  release; it references several v1 field names (`agent.ID`, `policy.ID`)
  that need a UUID-aware rewrite. Tracking this as a follow-up PR — the
  default `go test ./...` and `go vet ./...` both pass cleanly.

## [1.0.0] - 2025-12-29

### Added
- Complete Go SDK for Roe AI API
- Agent management (create, list, retrieve, update, delete, duplicate)
- Agent version management
- Job execution (sync and async)
- Batch job processing with `RunMany`
- File upload support (path, URL, bytes, reader)
- Context-aware operations for cancellation support
- Comprehensive error handling with typed errors
- Automatic retry logic with exponential backoff
- Request/response hooks for monitoring
- Pagination support for list operations
- Reference file downloads (screenshots, HTML, markdown)

### Changed
- Reorganized repository structure for better maintainability
- Renamed files for consistency (agents_api -> agents, http_client -> http)
- Merged model files into single types.go

### Documentation
- Complete README with examples
- 14 example programs demonstrating all features
- Inline documentation for all public APIs
