# Plan: Data Models for UniFi Resource Types

**Plan ID reference:** 2026-05-25-unifi-bootstrapper, Task 6.2
**PR:** Data models with JSON round-trip tests and golden fixtures
**Created:** 2026-05-25
**Predecessor:** [initial-scaffold.md](initial-scaffold.md)

## What

Data models representing all UniFi resource types that the
`unifi-bootstrapper` tool will enumerate from the UDM-Pro API:

- **APIResponse[T]** — generic API response envelope
- **FlexInt** — custom type handling JSON values that arrive as either
  strings or integers (firmware version variance)
- **Network** — VLAN, DHCP, and layer-2/3 configuration
- **FirewallRule** — traffic matching and action rules
- **FirewallGroup** — address-group and port-group collections
- **WLAN** — wireless network (SSID) configuration
- **PortForward** — port forwarding / NAT rules
- **PortProfile** — switch port VLAN/PoE/speed configuration
- **StaticRoute** — custom routing table entries
- **Device** — inventory-only device records (not Terraform-managed)
- **Inventory** — aggregate of all resource types with timestamp

Each model has a corresponding test file with JSON fixture loading,
round-trip marshal/unmarshal, omitempty verification, and edge case
coverage. Golden fixture files in `testdata/fixtures/` contain realistic
fake UDM-Pro API responses.

## Why

Models are the foundation of the entire tool. The client, exporter,
renderer, and generator packages all depend on these types to
deserialize API responses, manipulate data, and produce output. Getting
the JSON tags, field types, and serialization behavior right at this
stage prevents cascading bugs in every downstream package.

The `FlexInt` type specifically addresses a known firmware variance
where `rule_index` is returned as `"2000"` (string) on some firmware
versions and `2000` (integer) on others. Without this, the tool would
fail on a subset of UDM-Pro deployments.

## Context

Part of the Homelab & Home Automation project, Plan ID
`2026-05-25-unifi-bootstrapper`. This is Task 6.2 in the TDD
implementation sequence:

1. Project scaffold (complete, see `initial-scaffold.md`)
2. **This PR** — Data models and tests
3. API client
4. Exporter (resource enumeration)
5. Renderer (JSON/Markdown output)
6. Generator (Terraform file generation)
7. CLI wiring and integration tests

## Lessons Learned

- The UDM-Pro API uses inconsistent typing for `rule_index` across
  firmware versions. The `FlexInt` custom type with both int and string
  unmarshaling handles this cleanly without requiring firmware version
  detection.
- The `static-route_network` JSON key contains a hyphen, which is
  unusual for UniFi API fields. Go's `json` struct tags handle this
  without issue, but it warrants a dedicated test to catch any
  future regressions.
- Golden fixture files wrapped in the API response envelope
  (`{"meta":{"rc":"ok"},"data":[...]}`) allow testing the full
  deserialization path, not just individual model parsing.
