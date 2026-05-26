# Plan: JSON and Markdown renderers with passphrase redaction

**Plan ID:** 2026-05-25-unifi-bootstrapper
**Task:** 6.6
**Created:** 2026-05-25

## What

Implements two pure-function renderers in `internal/renderer/`:

- `RenderJSON(inv *models.Inventory) ([]byte, error)` -- produces indented
  JSON with all WLAN passphrases replaced by `"REDACTED"`.
- `RenderMarkdown(inv *models.Inventory) string` -- produces a
  human-readable Markdown document with tables for each resource type and
  all WLAN passphrases replaced by `REDACTED`.

Both renderers create a redacted copy of the inventory before producing
output so the caller's data is never mutated.

## Why

The `unifi-bootstrapper` CLI exports UniFi network configuration for
documentation and Terraform import. Exported inventories include WLAN
passphrases (`x_passphrase`), which are sensitive data that must never
appear in output files. These renderers provide safe-by-default output
formats that redact sensitive fields before any bytes leave the process.

JSON output supports machine-readable downstream consumption. Markdown
output supports human review and documentation.

## Context

- Builds on the models package (Task 6.2) which defines all resource
  types including `WLAN.XPassphrase`.
- The `internal/renderer/renderer.go` placeholder was replaced with the
  package-level doc comment.
- Tests use fake passphrase values from `testdata/fixtures/wlans.json`
  (`"test-passphrase-do-not-use"`, `"guest-fake-password-not-real"`) and
  verify redaction by scanning rendered output bytes for those known
  values.
- Predecessor plans: `docs/plans/02-models.md` (defines the data structures
  these renderers consume).

## Testing

- 12 unit tests covering both renderers.
- 98.5% statement coverage.
- Passphrase redaction tests confirm `REDACTED` is present and known
  passphrase strings are absent in both JSON and Markdown output.
- Input-mutation tests confirm the original inventory is never modified.
- Round-trip test confirms structural integrity after JSON
  marshal/unmarshal.
