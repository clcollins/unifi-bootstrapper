# Project Conventions

## Container Engine

- Primary engine: podman (preferred)
- Makefiles use a `CONTAINER_SUBSYS` variable defaulting to `podman`
  to allow overriding
- Never reference docker-specific tooling unless required for
  compatibility

## Language — Go

- Language: Go
- Go version: 1.26.3 (managed via gvm)
- Module path: `github.com/clcollins/unifi-bootstrapper`
- Entry point: `cmd/bootstrapper/main.go` using cobra + viper
- Linter: golangci-lint
- Test framework: `go test` with `testify` for assertions and mocks
- Build output: always to `/tmp/unifi-bootstrapper`, never into the
  repo directory

### Code Organization

```text
cmd/bootstrapper/     CLI entry point (cobra root command + subcommands)
internal/client/      HTTP client for UDM-Pro API
internal/models/      Data structures for UniFi resources
internal/exporter/    Resource fetcher (enumerates from API)
internal/generator/   Terraform file generator
internal/renderer/    JSON/Markdown inventory renderer
testdata/fixtures/    Mock API response fixtures for tests
```

### Go Style

- Follow standard Go conventions (effective Go, Go Code Review Comments)
- Use `internal/` for all non-CLI packages — nothing is public API
- Error handling: return errors, do not panic in library code
- Context: pass `context.Context` as first parameter where applicable
- Naming: use descriptive names; avoid single-letter variables outside
  loop indices

## Makefile Standards

- All targets `.PHONY`
- Must include `build`, `test`, `lint`, `clean`, `fmt`, `coverage`,
  `ci-checks`, and `ci-all` targets
- `ci-all` is the entry point for CI — runs all checks
- Variables for configurable values (Go binary path, build output path)
- Build output goes to `/tmp/`, not the repo directory

## CI Testing

### GitHub Actions

- CI runs on push to `main` and on pull requests to `main`
- CI job runs `make ci-all` which executes lint, test, and coverage
  serially
- Go version in CI matches `go.mod` toolchain version (1.26.3)

### Local vs Remote Execution

- Locally: `make ci-all` runs all checks serially
- Remotely: GitHub Actions runs `make ci-all` in a single job
- Both paths use the same Makefile targets, ensuring identical behavior

### Required Checks

| Check | Tool | Applies To |
|-------|------|------------|
| Go lint | golangci-lint | All Go source files |
| Go test | go test -race | All packages |
| Go vet | go vet (via golangci-lint) | All packages |
| Coverage | go test -coverprofile | All packages |

## Linting

- Fix all lint issues rather than suppressing rules, unless there is a
  documented reason
- Linter configuration: `.golangci.yml` at repo root
- Markdown linting: `.markdownlint.yaml` at repo root

## Documentation

### Plan Documents (in-repo)

- Every PR must have an associated plan document in `docs/plans/`
- Plan documents use descriptive filenames (e.g., `initial-scaffold.md`),
  not numeric prefixes
- Plans must consider lessons learned from previous plans in the same
  directory
- Superseded plans are preserved with a clear note at the top pointing
  to the replacement plan

### Markdown

- All Markdown should pass markdownlint (configured in
  `.markdownlint.yaml`)
- Use fenced code blocks with language identifiers
- Tables must have properly spaced separators
- Lists must be surrounded by blank lines

## Version Control

- Feature branches only; never commit directly to main
- DCO sign-off required on all commits (`-s` flag)
- Concise commit messages with descriptive body
- Attribution trailers required for AI-assisted work
- No force-push without explicit User approval
