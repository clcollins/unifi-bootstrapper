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

### golangci-lint v1 vs v2 with Go 1.26 (revised by PR #1)

The scaffold shipped `.golangci.yml` in v1 format and used
`golangci-lint-action@v6` with `version: latest`, which resolved to
v1.64.8 — built with Go 1.24. Go 1.26.3 targets fail with
`the Go language version used to build golangci-lint is lower than the
targeted Go version`. Fix: migrate config to v2 format (add `version:
"2"`, use `default: none`, drop `gosimple`/`typecheck` which are merged
or removed in v2) and pin the action to `v2.12.2`.

**Takeaway:** When scaffolding a Go project targeting a recent Go
version, always verify that golangci-lint's latest stable release was
built with a Go version >= the target. Pin the exact version in CI
rather than using `latest`, use the v2 config format, and use
`golangci-lint-action@v7` (v6 does not support golangci-lint v2).

### Parallel sub-agents need worktree isolation (Tasks 6.5 + 6.6)

Tasks 6.5 (Terraform generator) and 6.6 (renderers) were dispatched
as parallel sub-agents to maximize throughput. Both created separate
feature branches but shared the same local git clone. This caused
branch checkout conflicts — one agent's `git checkout` overwrote the
other's working tree state.

The work was recoverable because the agents touched completely
different packages (`internal/generator/` vs `internal/renderer/`)
and neither committed. The mixed working tree was sorted out manually
via stash-and-separate after both agents completed.

**Takeaway:** When dispatching parallel agents that create separate
git branches in the same repository, use worktree isolation (separate
git worktrees) or serialize the agents. Parallel dispatch into a
shared clone is only safe when agents work on the same branch or
modify non-overlapping files without switching branches.

### Plan document naming convention (revised after Task 6.6)

The original scaffold plan used descriptive filenames
(`initial-scaffold.md`, `models.md`, etc.) per the framework's
recommendation to avoid numeric prefixes since PR numbers aren't
known at creation time. In practice, this made it difficult to see
the chronological order of plans at a glance.

**Takeaway:** Use `##-summary.md` format (e.g., `01-initial-scaffold.md`,
`02-models.md`) for plan documents. The number reflects creation
order, not PR number. This provides natural sort order while keeping
descriptive names. Update cross-references when renaming.
