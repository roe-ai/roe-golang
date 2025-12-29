# Changelog

## 0.1.0 (unreleased)

- Added `ConfigParams` and expanded env parsing (`ROE_DEBUG`, proxy, extra headers, request IDs, retry tuning) with safer defaults.
- HTTP client now supports structured retries with backoff/jitter, proxy/connection pooling, request/response logging with redaction, and automatic request IDs.
- Expanded error taxonomy with rate-limit handling and request IDs/parsed detail on all API errors.
- `Job.WaitContext`/`JobBatch.WaitContext` add context cancellation, ordered batch handling, and clearer failure surfacing.
- File uploads now accept paths, readers/bytes, or URLs with validation and MIME sniffing; clearer errors for missing paths.
- Added unit tests for config parsing, HTTP retries, file handling, and batch ordering; integration tests are env-gated and CI workflow mirrors unit/integration split.
