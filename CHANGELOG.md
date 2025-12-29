# Changelog

All notable changes to the Roe AI Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
