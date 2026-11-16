# Changelog

All notable changes to the Roe Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.12] - 2026-11-11

Built from roe-main release `1-1-5`.

### Added

- Added the `ShieldEngage` value to `generated.ConnectorTypeEnum`, so you can now use `generated.ShieldEngage` anywhere this enum is accepted, including `generated.ConnectorMetadata`, `generated.CreateConnectionRequest`, and `generated.TestConnectionCredentialsRequest`.

## [1.3.11] - 2026-11-09

Built from roe-main release `1-1-5`.

### Added

- `generated.ConnectionsListParams` now supports an optional `Lifecycle` query parameter for `ConnectionsAPI.List`/`ListWithContext`, and `generated.Connection` plus `generated.ConnectionList` now include an optional `Lifecycle` field typed as `generated.LifecycleEnum`.
- Added `generated.LifecycleEnum` with the supported lifecycle values for connection filtering and connection payloads.
- `generated.UpdateConnectionRequest` and `generated.PatchedUpdateConnectionRequest` now include optional `ConfirmInUse` and `Lifecycle` fields, so callers can send lifecycle changes and confirm updates that affect connections already in use.
- Update and partial-update 409 responses now use `generated.ConnectionUpdateConflictResponse`, exposed as `JSON409` on `generated.ConnectionsUpdateResponse` and `generated.ConnectionsPartialUpdateResponse`.
- `generated.ListAgentJob` now includes the required `AgentVersionId` field.

## [1.3.10] - 2026-10-31

Built from roe-main release `1-1-4`.

### Added

- Added `Alloy` as a valid `generated.ConnectorTypeEnum` value, so you can now use `generated.Alloy` anywhere `ConnectorTypeEnum` is accepted, including `generated.CreateConnectionRequest.ConnectorType`, `generated.TestConnectionCredentialsRequest.ConnectorType`, and `generated.ConnectorMetadata.ConnectorType`.

## [1.3.9] - 2026-10-27

Built from roe-main release `1-1-3`.

### Added

- Added workflow management and execution wrappers on `client.Workflows`: `Create`, `CreateVersion`, `Delete`, `List`, `ListRuns`, `ListVersions`, `Retrieve`, `RetrieveRun`, `RetrieveVersion`, `Run`, `RunBatch`, `SetCurrentVersion`, and `Update`. You can now create workflows and versions, list and retrieve workflows, versions, and runs, execute workflows singly or in batches, switch the current version, and update or delete workflows through the typed client.
- Added raw workflow operations for the same API surface: `workflows_create`, `workflows_destroy`, `workflows_list`, `workflows_partial_update`, `workflows_retrieve`, `workflows_runs_batch_create`, `workflows_runs_create`, `workflows_runs_list`, `workflows_runs_retrieve`, `workflows_versions_create`, `workflows_versions_list`, `workflows_versions_retrieve`, and `workflows_versions_set_current_create`. These cover `POST /v1/workflows/`, `DELETE /v1/workflows/{workflow_id}/`, `GET /v1/workflows/`, `PATCH /v1/workflows/{workflow_id}/`, `GET /v1/workflows/{workflow_id}/`, `POST /v1/workflows/{workflow_id}/runs/batch/`, `POST /v1/workflows/{workflow_id}/runs/`, `GET /v1/workflows/{workflow_id}/runs/`, `GET /v1/workflows/runs/{run_id}/`, `POST /v1/workflows/{workflow_id}/versions/`, `GET /v1/workflows/{workflow_id}/versions/`, `GET /v1/workflows/{workflow_id}/versions/{version_id}/`, and `POST /v1/workflows/{workflow_id}/versions/{version_id}/set-current/`.
- Added workflow request, response, and pagination models in `generated`, including `Workflow`, `WorkflowDetail`, `WorkflowCreateRequest`, `WorkflowDefinition`, `WorkflowDefinitionRequest`, `WorkflowRun`, `WorkflowRunDetail`, `WorkflowRunBatch`, `WorkflowRunCreateRequest`, `WorkflowRunBatchCreateRequest`, `WorkflowVersion`, `WorkflowVersionDetail`, `WorkflowVersionCreateRequest`, `PatchedWorkflowUpdateRequest`, `PaginatedWorkflowList`, `PaginatedWorkflowRunList`, and `PaginatedWorkflowVersionList`. Use these types with the new workflow endpoints for typed create, update, list, versioning, and run operations.
- Added workflow definition and run component models in `generated`, including `AgentStep`, `ApprovalBlock`, `Condition`, `ConnectionStep`, `EditFieldsStep`, `ForEachBlock`, `IfBlock`, `InputField`, `MergeStep`, `ParallelBlock`, and `StepRun`. These support typed workflow definitions and detailed run inspection, along with the related enums `CombineEnum`, `GateKind`, `ModeEnum`, `OperatorEnum`, `StepRunDeliveryEnum`, `StepRunStatusEnum`, and `WorkflowRunStatusEnum`.

## [1.3.8] - 2026-10-20

Built from roe-main release `1-1-2`.

### Added

- Added the `ShieldApi` value to `generated.ConnectorTypeEnum`, so you can now use `generated.ShieldApi` anywhere this enum is accepted, including `generated.ConnectorMetadata`, `generated.CreateConnectionRequest`, and `generated.TestConnectionCredentialsRequest`.

## [1.3.7] - 2026-10-17

Built from roe-main release `1-1-2`.

### Added

- Added `Slack` and `Smtp` as valid `generated.ConnectorTypeEnum` values, so you can now use `generated.Slack` and `generated.Smtp` anywhere connector types are accepted, such as `generated.CreateConnectionRequest.ConnectorType` and `generated.TestConnectionCredentialsRequest.ConnectorType`.

## [1.3.6] - 2026-10-16

Built from roe-main release `1-1-1`.

### Added

- Added the `generated.SiftApi` value to `generated.ConnectorTypeEnum`, so you can now use `sift_api` anywhere this enum is accepted, including connector-related fields such as `ConnectorMetadata`, `CreateConnectionRequest`, and `TestConnectionCredentialsRequest`.

## [1.3.5] - 2026-10-13

Built from roe-main release `1-1-1`.

### Added

- `generated.AgentsJobsListParams` gains an optional `ReviewState` field for the `agents_jobs_list` query parameter `review_state`, with the new enum `generated.ReviewStateEnum` for its values. Use it through the generated client; the `client.Agents` wrapper does not expose this filter yet.
- `generated.ListAgentJob` now includes the `ReviewState` field, and `generated.SkillSetGenerationRun` now includes the `Error` field so callers can read a job's review state and any reported run failure reason from responses.

## [1.3.4] - 2026-10-08

Built from roe-main release `1-1-0`.

### Added

- Added a new `client.SkillSets` API with lifecycle, versioning, and generation methods: `Create`, `List`, `Retrieve`, `Update`, `Delete`, `ListVersions`, `RetrieveVersion`, `CreateVersion`, `SetCurrentVersion`, `GenerateVersion`, `GetGeneration`, and `CancelGeneration`. You can now manage skill sets and their versions through typed wrappers and the corresponding `generated` models such as `SkillSet`, `SkillSetVersion`, `CreateSkillSetRequest`, `CreateSkillSetVersionRequest`, `GenerateSkillSetVersionRequest`, and `GenerateSkillSetVersionResponse`.
- Added raw generated operations for the skill set endpoints: `skill_sets_create`, `skill_sets_destroy`, `skill_sets_list`, `skill_sets_retrieve`, `skill_sets_partial_update`, `skill_sets_versions_create`, `skill_sets_versions_list`, `skill_sets_versions_retrieve`, `skill_sets_set_current_version`, `skill_sets_generate_version`, `skill_sets_generation_retrieve`, and `skill_sets_generation_cancel`. If you call generated operations directly, you can now access the full `/v1/skills/` API surface.
- Added `TagsMatch` to `generated.AgentsListParams`, with enum values `All` and `Any`, for the `agents_list` query parameter `tags_match`. Callers can now control how `Tags` are matched when listing agents.
- Added connection authentication and browser-session models: `ConnectionAuthenticationCheck`, `ConnectionAuthenticationCheckStatusEnum`, `ConnectionAuthenticationChecks`, and `ConnectorBrowserSession`. These support the new `Checks` field on `TestConnection` and `BrowserSession` metadata on connectors.
- Added post-action and skill-set support models in `generated`, including `PostActionSpec`, `PostActionSpecRequest`, `SkillFile`, `SkillFileRequest`, `SkillGenerationStateEnum`, `PublicSkillSetGeneration`, `SkillSetDeleteConflict`, `SkillSetDependentAgentInfo`, `SkillSetGenerationRun`, `SkillSetVersionCreatedBy`, `PaginatedSkillSetList`, and `PaginatedSkillSetVersionList`. These types back the new skill set APIs and agent post-action fields.
- Added `Private` to `generated.AgentTag`, `PostActions` to `generated.AgentVersion`, and optional `PostActions` to `generated.AgentVersionCreateRequest`. You can now read configured post-actions on agent versions and send them when creating a new version.
- Added `SkillGenerationState` to `generated.BaseAgent`. Agent reads can now include the skill generation state.
- Added `BrowserSessionExpiresAt` to `generated.Connection`, `BrowserSession` to `generated.ConnectorMetadata`, and optional `Checks` to `generated.TestConnection`. Connection and connector responses can now include browser-session details and structured authentication check results.
- Added `Sift` and `Metabase` to `generated.ConnectorTypeEnum`. You can now use these values anywhere `ConnectorTypeEnum` is accepted.

### Changed

- **Breaking.** `AgentJobResultItem.Cost`, `AgentJobResultItem.InputTokens`, `AgentJobResultItem.OutputTokens`, `AgentJobResultResponse.InputTokens`, and `AgentJobResultResponse.OutputTokens` are no longer guaranteed to be present. Update code that dereferences these fields to handle `nil` values.

### Removed

- **Breaking.** Removed the `client.KnowledgeBase` API and its methods: `Catalog`, `Create`, `Delete`, `Finalize`, `ImportLens`, `LensByAtlasId`, `List`, `PatchSelection`, `PollDraft`, `Regenerate`, `Resolve`, `Retrieve`, `Sync`, and `Unlink`. Code that referenced `client.KnowledgeBase` will no longer compile and must be migrated off these wrappers.
- **Breaking.** Removed raw generated knowledge base operations: `knowledge_base_catalog_retrieve`, `knowledge_base_create`, `knowledge_base_destroy`, `knowledge_base_draft_retrieve`, `knowledge_base_finalize_create`, `knowledge_base_import_lens_create`, `knowledge_base_lens_retrieve`, `knowledge_base_list`, `knowledge_base_regenerate_create`, `knowledge_base_resolve_create`, `knowledge_base_retrieve`, `knowledge_base_selection_partial_update`, `knowledge_base_sync_create`, and `knowledge_base_unlink_destroy`. Direct callers of these generated endpoints must remove or replace those calls.
- **Breaking.** Removed knowledge-base-related models from `generated`, including `CreateKnowledgeBase`, `CreateKnowledgeBaseRequest`, `Draft`, `DraftRef`, `DraftStatusEnum`, `FinalizeRequest`, `KnowledgeBase`, `KnowledgeBaseStatusEnum`, `PaginatedKnowledgeBaseList`, `PatchedPatchSelectionRequest`, `PendingProposal`, `RegenerateRequest`, `RelevanceEnum`, and `ResolveRequest`. Any code importing or constructing these types must be updated.

## [1.3.3] - 2026-08-29

Built from roe-main release `1-0-95`.

### Added

- `generated.AgentVersion` now includes `SupportsMemory`, indicating whether the engine has a built-in memory profile and can be configured with memory profiles.
- `generated.ConnectorTypeEnum` now accepts the `Shield` value, which can be used anywhere this enum is sent or received, including connector-related request and metadata types.
- `generated.SupportedLLMModel` now includes the required `SupportsThinking` field so callers can detect whether a model supports thinking.

## [1.3.2] - 2026-08-18

Built from roe-main release `1-0-93`.

### Added

- `client.Agents.Jobs.ResendWebhook` and `ResendWebhookWithContext` accept an optional trailing `webhookID ...string`, so you can resend a job's completion webhook to one specific webhook instead of all of the agent's webhooks. Existing calls still compile.

## [1.3.1] - 2026-08-18

Built from roe-main release `1-0-93`.

### Added

- Added support for re-sending an agent job webhook with `client.Agents.Jobs.ResendWebhook`, plus the raw generated operation `AgentsJobsWebhookResendCreate`/`AgentsJobsWebhookResendCreateWithResponse` for `POST /v1/agents/jobs/{job_id}/webhook/resend/`. Use `generated.ResendAgentJobWebhookRequest` as the request body and read the result from `generated.AgentJobWebhookResendResponse` (or `AgentsJobsWebhookResendCreateResponse.JSON200` on the response client).

## [1.3.0] - 2026-08-06

> ### ⚠️ This release contains source-breaking changes
>
> Existing code will not compile after upgrading. Go's minimal version
> selection means a pinned `go.mod` will not move on its own — you are only
> affected when you explicitly `go get -u`. Both breaks and their migrations
> are listed under **Changed** below.
>
> This is a minor bump rather than a major one because the module path
> `github.com/roe-ai/roe-golang` has no `/v2` suffix, and introducing one
> would force every consumer to rewrite imports for what amounts to one extra
> argument. Strict SemVer would call this 2.0.0; the trade was made
> deliberately.

Catch-up release covering roe-main `1-0-88` through `1-0-91`. The Go SDK was
stalled at `1-0-87` for four releases while the wrapper generator panicked on
the `map[string]string` contract type (fixed in #38), so this bundles the
whole gap into one release. The equivalent Python and TypeScript changes ship
separately through their own release PRs.

### Added
- `dynamicInputs map[string]string` on `Connections.Create`, `Update`,
  `Replace`, and `TestCredentials` (and their `WithContext` variants), for
  connector-level dynamic runtime inputs. Omitted from the request body when
  empty — note that an allocated-but-empty map is also omitted, so this
  cannot be used to clear existing dynamic inputs.
- `CredentialsConfigured`, `DynamicInputs`, and
  `DynamicInputTestDisabledReason` on the connection response types.
- `DynamicInputFields` and `DynamicInputTestFields` on `ConnectorMetadata`.
  These are required, non-pointer fields.
- `UpdatedAt` on the agent response type, and `UpdatedFrom` / `UpdatedTo`
  filters on `AgentsListParams`.
- `ordering` query parameter on the connections list endpoint. Available on
  the generated client only — the friendly `Connections.List` wrapper does
  not expose it, because `openapi/wrappers.yml` was not updated for that
  operation upstream.

### Changed
- **Breaking.** The four `Connections` methods above gained a trailing
  positional `dynamicInputs` parameter. Unlike the variadic `RunOptions`
  change in 1.2.0, this does not compile unchanged.

  ```go
  // before
  client.Connections.Update(id, name, desc, config, authConfig)
  // after — nil preserves the previous behavior exactly
  client.Connections.Update(id, name, desc, config, authConfig, nil)
  ```

### Removed
- **Breaking.** `AgentJobStatusEvent` is renamed `PublicAgentJobStatusEvent`
  to match the spec component name. The only place it was reachable is
  `ListAgentJob.StatusEvents`, whose element type changes accordingly.

## [1.2.0] - 2026-07-14

### Added
- `RunOptions` variadic options struct on `Agents.Run`, `RunMany`, `RunSync`,
  `RunVersion`, `RunVersionSync` (and their `WithContext` variants), plus the
  object-style `BaseAgent.Run` and `AgentVersion.Run` helpers.
  `roe.RunOptions{SkipCache: true}` sends the `X-Skip-Cache: true` header so
  the backend bypasses the job-result cache and forces a fresh run (the fresh
  result still refreshes the cache). Also re-exported from the deprecated
  `roe` subpackage shim.

### Changed
- The run-method signatures above gained a trailing variadic
  `opts ...RunOptions` parameter. Ordinary call sites compile unchanged, but
  code that stores these methods in `func`-typed variables or declares
  interfaces with the exact old signatures must be updated. Release this as a
  minor version bump, not a patch.
- Per-request headers now override same-named `Config.ExtraHeaders` entries
  instead of sending duplicate header lines (a duplicated `X-Skip-Cache`
  would be folded to `"true,true"` and ignored by the backend).

## [1.1.5] - 2026-07-08

Built from roe-main release `1-0-87`.

### Added

- Added agent job listing support via `client.Agents.Jobs.ListJobs` and the raw operation `agents_jobs_list` for `GET /v1/agents/{agent_id}/jobs/`. Callers can now list an agent’s jobs and use the generated request/response types `generated.AgentsJobsListParams` and `generated.PaginatedListAgentJobList`.
- Added generated models for agent job list results and related data: `generated.ListAgentJob`, `generated.AgentJobEvaluation`, `generated.AgentJobFeedbackNested`, `generated.AgentJobStatusEvent`, and `generated.JobInput`. These types expose job evaluation, review feedback, status history, and input payload details returned by job listing.
- Added `generated.ReviewStatusEnum` for review state values used by `generated.AgentJobFeedbackNested.ReviewStatus`. You can now work with typed review statuses instead of raw strings.

## [1.1.4] - 2026-07-04

Built from roe-main release `1-0-87`.

### Added

- `generated.AgentsListParams` gains optional filters for the raw `agents_list` operation: `CreatedFrom`, `CreatedTo`, `CreatorId`, `IncludeUntagged`, `JobCountMax`, `JobCountMin`, `MostRecentJobFrom`, `MostRecentJobTo`, and `Name`. Use them through the generated client; `client.Agents.List` still takes only `page` and `pageSize`.

## [1.1.3] - 2026-07-03

Built from roe-main release `1-0-86`.

### Added

- Added `client.Agents.Jobs.RetrieveArtifact` and the raw operation `agents_jobs_artifacts_result_retrieve` (`GET /v1/agents/jobs/{agent_job_id}/artifacts/result/`) so you can fetch a job artifact result; this also adds `generated.AgentJobArtifactResult` for the response.
- Added the `client.KnowledgeBase` API with `Catalog`, `Create`, `Delete`, `Finalize`, `ImportLens`, `LensByAtlasId`, `List`, `PatchSelection`, `PollDraft`, `Regenerate`, `Resolve`, `Retrieve`, `Sync`, and `Unlink`, plus the corresponding raw operations for `/v1/knowledge-base/` and related routes.
- Added knowledge base models in `generated`, including `CreateKnowledgeBase`, `CreateKnowledgeBaseRequest`, `KnowledgeBase`, `PaginatedKnowledgeBaseList`, `Draft`, `DraftRef`, `PendingProposal`, `FinalizeRequest`, `PatchedPatchSelectionRequest`, `RegenerateRequest`, `ResolveRequest`, `DraftStatusEnum`, `KnowledgeBaseStatusEnum`, and `RelevanceEnum` for the new `client.KnowledgeBase` workflows.
- Added `generated.AgentJobCancelAllResponse`.
- Added `google_docs` and `google_sheets` to `ConnectorTypeEnum`; in Go these are exposed as `generated.GoogleDocs` and `generated.GoogleSheets`.

### Removed

- **Breaking.** Removed the `organization_id` query parameter from the raw `agents_jobs_references_retrieve` operation. If you set `OrganizationId` on `generated.AgentsJobsReferencesRetrieveParams`, remove it.

## [1.1.2] - 2026-07-02

Built from roe-main release `1-1-1`.

### Added

- Added `client.Agents.Replace`, `client.Agents.Versions.Replace`, and `client.Policies.Replace` so you can fully replace existing agents, agent versions, and policies through the SDK.
- Added connection management methods on `client.Connections`: `List`, `Create`, `Retrieve`, `Update`, `Replace`, `Delete`, `Test`, and `TestCredentials`. You can now manage saved connections and validate credentials, with responses using generated types including `generated.PaginatedConnectionListList`, `generated.Connection`, and `generated.TestConnection`.
- Added connector discovery methods `client.Connectors.List` and `client.Connectors.Retrieve` to list available connector types and fetch connector metadata, returning `generated.ConnectorListResponse` and `generated.ConnectorMetadata`.
- Expanded `client.Tables` with `List`, `Describe`, `Preview`, `Query`, `QueryResult`, and `Delete`. You can now inspect tables, preview rows, submit read-only queries and fetch their results, and delete tables using generated response types such as `generated.TableListResponse`, `generated.TableDescribeResponse`, `generated.TablePreviewResponse`, `generated.TableQuerySubmitResponse`, and `generated.TableQueryResultResponse`.
- Added `client.Users.Me` to retrieve the currently authenticated user as `generated.User`.

## [1.1.1] - 2026-06-18

Built from roe-main release `1-1-1`.

### Added

- Added the `client.Agents` API with lifecycle and execution methods: `client.Agents.Create`, `client.Agents.Delete`, `client.Agents.Duplicate`, `client.Agents.List`, `client.Agents.Retrieve`, `client.Agents.Update`, `client.Agents.Run`, `client.Agents.RunMany`, `client.Agents.RunSync`, `client.Agents.RunVersion`, and `client.Agents.RunVersionSync`.
- Added `client.Agents.Jobs` methods to manage agent runs and results: `client.Agents.Jobs.Cancel`, `client.Agents.Jobs.CancelAll`, `client.Agents.Jobs.DeleteData`, `client.Agents.Jobs.DownloadReference`, `client.Agents.Jobs.RetrieveResult`, `client.Agents.Jobs.RetrieveResultMany`, `client.Agents.Jobs.RetrieveStatus`, and `client.Agents.Jobs.RetrieveStatusMany`.
- Added the `client.Agents.Versions` API with version management methods: `client.Agents.Versions.Create`, `client.Agents.Versions.Delete`, `client.Agents.Versions.List`, `client.Agents.Versions.ListPaginated`, `client.Agents.Versions.Retrieve`, `client.Agents.Versions.RetrieveCurrent`, `client.Agents.Versions.RetrieveCurrentWithEval`, and `client.Agents.Versions.Update`.
- Added the `client.Policies` API with resource methods `client.Policies.Create`, `client.Policies.Delete`, `client.Policies.List`, `client.Policies.Retrieve`, and `client.Policies.Update`, plus version methods `client.Policies.Versions.Create`, `client.Policies.Versions.List`, and `client.Policies.Versions.Retrieve`.

## [1.1.0] - 2026-06-12

Built from roe-main release `1-0-83`.

### Added

- Added raw client operations for connections, connectors, and tables: `ConnectionsCreate`, `ConnectionsDestroy`, `ConnectionsList`, `ConnectionsPartialUpdate`, `ConnectionsRetrieve`, `ConnectionsTestCreate`, `ConnectionsTestCredentialsCreate`, `ConnectionsUpdate`, `ConnectorsRetrieve`, `ConnectorsRetrieveByType`, `TablesDescribeRetrieve`, `TablesDestroy`, `TablesList`, `TablesPreviewRetrieve`, `TablesQueryCreate`, and `TablesQueryResultRetrieve`. These operations come with new `generated` request and response types including `Connection`, `ConnectionList`, `CreateConnectionRequest`, `UpdateConnectionRequest`, `PatchedUpdateConnectionRequest`, `TestConnection`, `TestConnectionCredentialsRequest`, `ConnectorListResponse`, `ConnectorMetadata`, `TableListResponse`, `TableDescribeResponse`, `TablePreviewResponse`, `TableQueryRequest`, `TableQuerySubmitResponse`, and `TableQueryResultResponse`.
- `generated.AgentsJobsReferencesRetrieveParams` now includes the optional `Download` query parameter, so callers can request `agents_jobs_references_retrieve` with `download=true`.
- Added new `generated` models and enums used by agent, policy, connection, and error responses, including `AgentExecutionRequest`, `AgentJobResultManyRequest`, `AgentJobSingleStatus`, `AgentJobStatusManyRequest`, `AgentRunAsyncManyRequest`, `AgentTag`, `AgentVersionUpdateRequest`, `PatchedAgentVersionUpdateRequest`, `ApiErrorResponse`, `ErrorDetailResponse`, `MessageResponse`, `DependentAgentInfo`, `PolicyDeleteConflict`, `QdrantCleanupErrorResponse`, `DuplicateConnectionExisting`, `DuplicateConnectionResponse`, `ConnectionDeleteErrorResponse`, `PaginatedConnectionListList`, `User`, `ColorEnum`, `ConnectorTypeEnum`, `StatusEnum`, and `TableQueryStatusEnum`.

### Changed

- **Breaking.** `generated.AgentJobDeleteDataResponse.Errors` changed from an optional `[]interface{}`-style field to a required nullable string list (`*[]string`). Update code that reads or constructs this field to handle `nil` and string values only.
- **Breaking.** `generated.AgentJobStatus` changed shape: `ErrorMessage` is now nullable, `Status` is now nullable, `Timestamp` is now an optional `float64`, and the model now also includes required `CreatedAt`, `Id`, and `LastUpdatedAt`. Code consuming job status responses must handle missing or `nil` values and the new fields.
- **Breaking.** `generated.AgentVersionCreateRequest.EngineConfig` and `generated.AgentVersionCreateRequest.InputDefinitions` are no longer guaranteed to be present, and `generated.BaseAgentCreateRequest.EngineConfig`, `InputDefinitions`, and `OrganizationId` are now optional as well. Code constructing these request bodies must stop assuming those fields are required.
- **Breaking.** `generated.BaseAgent.MostRecentJob` is now `*time.Time` instead of `*string`, and `generated.BaseAgent.Tags` is now `*[]generated.AgentTag` instead of `*string`. Update any parsing or direct field usage to work with typed timestamps and structured tags.
- **Breaking.** `generated.PolicyVersion.BaseVersionId` is now nullable, so callers must handle `nil` when reading this field.

### Removed

- **Breaking.** Removed the old generated model names `AgentExecutionRequestRequest`, `AgentJobResultManyRequestRequest`, `AgentJobStatusManyRequestRequest`, `AgentRunAsyncManyRequestRequest`, `AgentVersionUpdateRequestRequest`, and `PatchedPatchedAgentVersionUpdateRequestRequest`; update references to `AgentExecutionRequest`, `AgentJobResultManyRequest`, `AgentJobStatusManyRequest`, `AgentRunAsyncManyRequest`, `AgentVersionUpdateRequest`, and `PatchedAgentVersionUpdateRequest` instead.
- **Breaking.** Removed `generated.ErrorResponse` and `generated.PaginatedAgentJobResultItemList`. If you referenced these types directly, switch to the currently generated response models returned by the relevant operations.

## [1.0.803] - 2026-06-05

Built from roe-main release `1-0-82`.

### Changed

- Regenerated from the latest SDK spec; no changes to the generated API surface.

## [1.0.802] - 2026-05-22

### Added
- Generated friendly wrappers for discovery and table upload:
  `client.Discovery.ListAgentEngineTypes(...)`,
  `client.Discovery.ListSupportedModels(...)`, and `client.Tables.Upload(...)`.

### Changed
- Versions are now synchronized across roe-python (`roe-ai`), roe-typescript,
  and roe-golang. The public SDKs share a single 1.0.x patch counter driven by
  the SDK OpenAPI spec via the roe-main release pipeline.

## [1.0.0] - 2025-12-29

### Added
- Complete Go SDK for Roe API
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
