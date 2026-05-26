# Plan: HTTP Client for UDM-Pro API Communication

**Plan ID reference:** 2026-05-25-unifi-bootstrapper, Task 6.3
**PR:** HTTP client with API key and cookie-session auth
**Created:** 2026-05-25
**Predecessor:** [models.md](models.md)

## What

HTTP client implementing `ClientInterface` with two authentication
modes for connecting to a UDM-Pro local API:

- **ClientInterface** -- defines the contract for all resource-fetching
  methods (Ping, GetNetworks, GetFirewallRules, GetFirewallGroups,
  GetWLANs, GetPortForwards, GetPortProfiles, GetStaticRoutes,
  GetDevices)
- **Client** -- real HTTP client using functional options pattern for
  configuration (API key auth, cookie-session auth, insecure TLS,
  custom HTTP client injection)
- **MockClient** -- testify/mock implementation of ClientInterface for
  use in downstream package tests (exporter, generator, renderer)

All API paths are prefixed with `/proxy/network` as required by the
UDM-Pro gateway proxy architecture.

## Why

The HTTP client is the sole point of contact between the tool and the
UDM-Pro API. Centralizing authentication, request construction, error
handling, and response deserialization here means all downstream
packages (exporter, generator, renderer) work exclusively with typed Go
structs and a mockable interface. This separation enables comprehensive
testing without any real network dependencies.

Two authentication modes are necessary because:

- **API key auth** is simpler and preferred for automation (set a single
  header on every request)
- **Cookie-session auth** is the legacy mechanism some deployments
  require (POST login, capture session cookies, extract CSRF token for
  subsequent requests)

## Context

Part of the Homelab & Home Automation project, Plan ID
`2026-05-25-unifi-bootstrapper`. This is Task 6.3 in the TDD
implementation sequence:

1. Project scaffold (complete, see `initial-scaffold.md`)
2. Data models and tests (complete, see `models.md`)
3. **This PR** -- HTTP client with dual auth and MockClient
4. Exporter (resource enumeration)
5. Renderer (JSON/Markdown output)
6. Generator (Terraform file generation)
7. CLI wiring and integration tests

## Design Decisions

- **Functional options pattern** (`WithAPIKey`, `WithCredentials`,
  `WithInsecure`, `WithHTTPClient`) for clean, extensible configuration
  without a proliferating constructor parameter list.
- **Generic `fetch[T]` helper** uses Go generics to eliminate
  boilerplate across all eight resource-fetching methods. Each public
  method is a one-liner that calls `fetch` with the appropriate type
  parameter and endpoint path.
- **`apiErrorResponse` internal struct** captures the `meta.msg` field
  from error responses, which the shared `models.APIResponse` type does
  not expose. This avoids polluting the models package with
  error-handling concerns specific to the client.
- **Cookie jar on all clients** ensures session cookies are automatically
  managed regardless of auth mode, simplifying the implementation.
- **`httptest.NewTLSServer`** used exclusively in tests to verify TLS
  behavior without real network calls.

## Lessons Learned

- The UDM-Pro API returns CSRF tokens in a cookie named `TOKEN`. The
  value must be extracted and sent as an `X-CSRF-Token` header on all
  subsequent requests after login. This is not documented in any
  official API reference and was determined empirically.
- `errcheck` linter flags `defer resp.Body.Close()` as an unchecked
  error. The idiomatic Go fix is
  `defer func() { _ = resp.Body.Close() }()` which explicitly discards
  the error, satisfying the linter while acknowledging that a Close
  error on a read-only response body is not actionable.
- Testing connection failures with unreachable IP addresses
  (e.g., `192.0.2.1`) causes TCP timeout delays that slow the test
  suite. Using a closed `httptest` server produces an immediate
  connection-refused error, which is faster and equally valid.
