# Plan: Initial Project Scaffold

**Plan ID reference:** 2026-05-25-unifi-bootstrapper
**PR:** Initial scaffold — directory structure, build system, CI, docs
**Created:** 2026-05-25

## What

Initial project scaffold for the `unifi-bootstrapper` Go CLI tool.
Creates:

- Go module with cobra, viper, and testify dependencies
- Directory structure: `cmd/bootstrapper/`, `internal/client/`,
  `internal/models/`, `internal/exporter/`, `internal/generator/`,
  `internal/renderer/`, `testdata/fixtures/`
- Makefile with build, test, lint, coverage, fmt, ci-checks, ci-all,
  and clean targets
- GitHub Actions CI workflow running `make ci-all`
- golangci-lint and markdownlint configuration
- Documentation triad: `CLAUDE.md`, `AGENTS.md`, `CONVENTIONS.md`
- README with project overview, prerequisites, and usage
- Apache 2.0 LICENSE
- `.gitignore` covering build output, generated files, and sensitive
  output directory

## Why

Foundation for TDD development of the unifi-bootstrapper CLI tool. All
subsequent implementation work depends on having a valid Go module,
build system, CI pipeline, and established conventions. This is the
first PR in a 7-PR TDD sequence that will build up the tool
incrementally:

1. **This PR** — Project scaffold
2. Models and client — data structures and API client
3. Exporter — resource enumeration
4. Renderer — JSON/Markdown output
5. Generator — Terraform file generation
6. CLI wiring — cobra commands connecting all packages
7. Integration tests and polish

## Context

Part of the Homelab & Home Automation project. The unifi-bootstrapper
tool addresses the problem of bootstrapping Terraform management for an
existing UniFi network deployment. Rather than manually discovering and
importing each resource, this tool will connect to the UDM-Pro API,
enumerate all Terraform-manageable resources, and produce ready-to-use
import blocks.

This is the first plan document in this repository — no predecessors to
reference.

## Lessons Learned

*No prior plans exist — this section will be populated as the project
progresses and subsequent plans reference lessons from this scaffold.*
