# Changelog

All notable changes to the Roe Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
