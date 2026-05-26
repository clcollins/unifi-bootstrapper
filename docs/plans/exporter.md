# Plan: Concurrent resource exporter

**Plan ID:** 2026-05-25-unifi-bootstrapper, Task 6.4
**Author:** Implementer agent
**Status:** Complete

## What

The `internal/exporter` package provides an `Exporter` type that
orchestrates concurrent fetches of all 8 UniFi resource types from the
UDM-Pro API via `ClientInterface`, aggregating results into a
`models.Inventory`.

## Why

Fetching 8 independent resource endpoints sequentially would take
roughly 8x the round-trip time. Concurrent fetches minimize total export
time while keeping the code simple. The exporter also handles partial
failures gracefully: if some endpoints succeed and others fail, the
caller receives both the partial inventory and a combined error,
allowing downstream consumers to decide whether partial data is
acceptable.

## Context

- Builds on `docs/plans/http-client.md` which established the
  `ClientInterface` abstraction and `MockClient` for testing.
- Builds on `docs/plans/models.md` which defined the `Inventory` struct
  aggregating all resource types.
- The exporter is a pure orchestrator: it knows nothing about HTTP,
  JSON, or the API wire format. It only calls `ClientInterface` methods
  and assembles results.

## Design decisions

- **Goroutines with sync.WaitGroup and mutex** over `errgroup`: The
  standard `errgroup` cancels on first error, which conflicts with the
  requirement to collect all errors and return partial results. Using
  raw goroutines with a WaitGroup and mutex provides the needed
  control.
- **errors.Join** for combining errors: Available since Go 1.20, this
  produces a clean multi-error value that preserves all wrapped errors
  for `errors.Is`/`errors.As` inspection by callers.
- **ExportedAt set after all fetches complete**: This timestamp
  represents when the snapshot was finalized, not when it started.

## Testing

Tests use `MockClient` (testify/mock) exclusively. Coverage is 100% of
statements. Test scenarios:

- All 8 endpoints succeed (happy path)
- Single endpoint failure (partial results returned)
- Multiple endpoint failures (partial results returned)
- All endpoints fail (combined error returned)
- Empty responses (non-nil empty slices preserved)
- Context cancellation (errors propagated)
- ExportedAt timestamp validation
