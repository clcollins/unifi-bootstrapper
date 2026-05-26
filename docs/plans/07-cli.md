# Plan: Cobra CLI entry point with ping, export, version subcommands

**Plan ID:** 2026-05-25-unifi-bootstrapper
**Task:** 6.7
**Created:** 2026-05-25

## What

Implements the CLI entry point for unifi-bootstrapper using Cobra for
command structure and Viper for flag/environment-variable binding. The
CLI wires together all internal packages (client, exporter, generator,
renderer) into three subcommands:

- **version** -- prints the build-time version string and exits
- **ping** -- tests connectivity and authentication against the UDM-Pro API
- **export** -- fetches all resources, renders JSON/Markdown inventory
  files, and generates Terraform provider/import/stub configurations

## Why

This is the final integration step that makes unifi-bootstrapper a
usable tool. All prior tasks (01-06) built internal packages; this task
composes them into a single binary with a user-facing CLI.

## Context

- **Plan ID:** 2026-05-25-unifi-bootstrapper, Task 6.7
- **Depends on:** Tasks 6.1 (scaffold), 6.2 (models), 6.3 (client),
  6.4 (exporter), 6.5 (generator), 6.6 (renderers)
- **Predecessors:**
  - `docs/plans/01-initial-scaffold.md` -- repo structure and Makefile
  - `docs/plans/02-models.md` -- data models
  - `docs/plans/03-http-client.md` -- API client
  - `docs/plans/04-exporter.md` -- concurrent resource fetcher
  - `docs/plans/05-terraform-generator.md` -- Terraform file generation
  - `docs/plans/06-renderers.md` -- JSON/Markdown output

## Design decisions

- **Cobra + Viper** for CLI framework per project convention (standing
  User preference for Go projects).
- **All flags have environment variable equivalents** via Viper with
  `UNIFI_` prefix, enabling container/CI use without flag passing.
- **Partial failure handling** in export: when some API endpoints fail
  but others succeed, the tool writes partial output files with warnings
  and exits non-zero. This is more useful than failing entirely.
- **Build-time version injection** via `-ldflags "-X main.version=..."`,
  with `dev` as the default for development builds.
- **testHTTPClient package var** allows integration tests to inject
  httptest server clients without modifying the client package's public
  API. This is scoped to the main package and not exported.

## Test coverage

Integration tests use `httptest.NewTLSServer` with fixture data from
`testdata/fixtures/` to exercise the full request path through the real
client, exporter, generator, and renderer packages:

- Version output verification
- Ping success and auth failure paths
- Export success with all 5 output file verification
- Export partial failure with warning output
- Missing host validation
- Help text content verification
- Viper environment variable binding
- Directory creation for nested output paths
