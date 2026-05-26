# unifi-bootstrapper

A Go CLI tool that connects to a UniFi UDM-Pro's local API, enumerates
Terraform-manageable resources, and emits JSON/Markdown inventory files
and Terraform import configuration. Designed to bootstrap Terraform
management of an existing UniFi deployment without manually enumerating
every resource.

## Prerequisites

- Go >= 1.26.3
- Terraform >= 1.5.0 (for [import blocks](https://developer.hashicorp.com/terraform/language/import))
- Access to a UniFi UDM-Pro with a local API key
- golangci-lint (for linting)

## Build

```bash
make build
```

Builds the binary to `/tmp/unifi-bootstrapper`.

## Test

```bash
make test
```

Runs all tests with race detection enabled.

## Lint

```bash
make lint
```

## Full CI

```bash
make ci-all
```

Runs lint, test, and coverage checks — the same targets CI executes.

## Usage

**Note:** Subcommands are not yet implemented. The following describes
the planned CLI interface.

```text
unifi-bootstrapper [command]

Available Commands:
  ping        Verify connectivity to the UDM-Pro API
  export      Enumerate resources and emit inventory + Terraform files
  version     Print the version
```

### Configuration

The tool reads configuration from environment variables or a config
file:

- `UNIFI_HOST` — UDM-Pro hostname or IP address
- `UNIFI_API_KEY` — API key for authentication
- `UNIFI_SITE` — UniFi site name (default: `default`)

### Output Files

| File | Description |
|------|-------------|
| `output/inventory.json` | Full resource inventory in JSON format |
| `output/inventory.md` | Human-readable Markdown inventory |
| `output/terraform/provider.tf` | Terraform provider configuration |
| `output/terraform/imports.tf` | Terraform import blocks for all discovered resources |

## Security

The `output/` directory contains sensitive network topology data
(device IPs, MAC addresses, VLAN configurations, firewall rules) and is
gitignored. Never commit its contents.

WLAN passphrases are explicitly filtered from all output files.

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.

---

Created with assistance from Claude <noreply@anthropic.com>
