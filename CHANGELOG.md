# Changelog

All notable changes to the Roe AI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-05-20

### Changed
- **Module path migration (BREAKING).** The Go module path is now
  `github.com/roe-ai/roe-golang/v2`. Customers must update their imports:

  ```diff
  - import "github.com/roe-ai/roe-golang"
  + import "github.com/roe-ai/roe-golang/v2"
  ```

  And reinstall:

  ```bash
  go get github.com/roe-ai/roe-golang/v2@v2.0.0
  ```

  This is required by Go's
  [module versioning rules](https://go.dev/doc/modules/version-numbers#v2-go-modules)
  for majors ≥ 2. No API surface changes vs. 1.0.80 — only the import path.

- Versions are now synchronized across roe-python (`roe-ai`), roe-typescript,
  roe-golang, and roe-mcp. All four packages share a single patch counter,
  driven by the SDK OpenAPI spec via the roe-main release pipeline.

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
