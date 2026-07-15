# Changelog

All notable changes to the Roe Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `RunOption` variadic options on `Agents.Run`, `RunMany`, `RunSync`,
  `RunVersion`, `RunVersionSync` (and their `WithContext` variants), plus the
  object-style `BaseAgent.Run` and `AgentVersion.Run` helpers.
  `roe.WithSkipCache(true)` sends the `X-Skip-Cache: true` header so the
  backend bypasses the job-result cache and forces a fresh run (the fresh
  result still refreshes the cache). Also re-exported from the deprecated
  `roe` subpackage shim.

### Changed
- The run-method signatures above gained a trailing variadic
  `opts ...RunOption` parameter. Ordinary call sites compile unchanged, but
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
