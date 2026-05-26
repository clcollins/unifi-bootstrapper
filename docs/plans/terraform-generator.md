# Terraform Generator

## What

Pure Terraform file generator functions that take a `models.Inventory`
and produce three HCL file contents as strings:

- **`provider.tf`** -- Pins the `filipowm/unifi` provider to `~> 1.0.0`
  and declares `unifi_api_url` and `unifi_api_key` variables.
- **`imports.tf`** -- One `import {}` block per non-device resource,
  mapping each resource to its Terraform type and sanitized name.
- **`stubs.tf`** -- Minimal `resource` skeletons for each non-device
  resource. WLAN stubs include a passphrase placeholder
  (`REPLACE_WITH_ACTUAL_PASSPHRASE`) to prevent real passphrases from
  appearing in generated output.

A `SanitizeName` helper converts arbitrary resource names to valid
Terraform identifiers (spaces/hyphens to underscores, special characters
removed, digit-leading names prefixed with underscore, empty names
defaulting to `unnamed`).

All functions are pure -- same input, same output, no side effects, no
file I/O. File writing is the CLI layer's responsibility (Task 6.7).

## Why

The unifi-bootstrapper tool needs to produce HCL files so users can
import existing UDM-Pro resources into Terraform state. The
`filipowm/unifi` provider supports `terraform plan -generate-config-out`
to flesh out stubs after import, so the generator only needs to produce
minimal skeletons alongside the import blocks.

Keeping generators pure makes them trivially testable and composable.
The CLI layer can write the strings to files, combine them, or pipe them
elsewhere without the generator knowing or caring.

## Context

- **Plan ID:** 2026-05-25-unifi-bootstrapper
- **Task:** 6.5 -- Implement Terraform import block and resource stub
  generator
- **Builds on:** models (Task 6.2, `docs/plans/models.md`) which
  defines `Inventory`, `Network`, `FirewallRule`, `FirewallGroup`,
  `WLAN`, `PortForward`, `PortProfile`, `StaticRoute`, and `Device`
  types.

## Resource type mapping

| Model type    | Terraform resource type  |
|---------------|--------------------------|
| Network       | `unifi_network`          |
| FirewallRule  | `unifi_firewall_rule`    |
| FirewallGroup | `unifi_firewall_group`   |
| WLAN          | `unifi_wlan`             |
| PortForward   | `unifi_port_forward`     |
| PortProfile   | `unifi_port_profile`     |
| StaticRoute   | `unifi_static_route`     |
| Device        | excluded (inventory-only)|

## Design decisions

- **`fmt.Sprintf` / `strings.Builder` over `text/template`:** The HCL
  output is simple enough that string formatting is clearer and faster
  than template machinery. Templates would add parsing/execution
  overhead with no readability benefit for this use case.
- **Devices excluded:** Devices are inventory-only in the
  `filipowm/unifi` provider -- there is no `unifi_device` resource type
  to import into.
- **WLAN passphrase placeholder:** WLAN stubs use
  `REPLACE_WITH_ACTUAL_PASSPHRASE` rather than the real `XPassphrase`
  value from the model to prevent accidental credential exposure in
  generated files.

## Test coverage

100% statement coverage via table-driven tests covering:

- `SanitizeName` edge cases (spaces, special chars, digit-leading,
  empty, already-valid)
- `GenerateProvider` content verification and golden-file comparison
- `GenerateImports` with full inventory, empty inventory, nil slices,
  devices-only, name sanitization, and golden-file comparison
- `GenerateStubs` with full inventory, empty inventory, nil slices,
  devices-only, WLAN passphrase redaction, and golden-file comparison
- Purity assertions (same input produces same output)

## Lessons from prior plans

- The models plan (`docs/plans/models.md`) established the convention
  of keeping packages pure and side-effect-free where possible. This
  generator follows the same principle.
- The renderer plan demonstrates the pattern of redacting sensitive
  WLAN passphrases in output, which this generator adopts for stubs.
