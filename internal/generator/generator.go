// Package generator produces Terraform configuration files including
// provider blocks, import blocks, and resource stubs from exported
// UniFi resources. All functions are pure — same input, same output,
// no side effects, no file I/O.
package generator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

// SanitizeName converts a resource name to a valid Terraform identifier.
// Spaces and hyphens become underscores, other non-alphanumeric/underscore
// characters are removed, names starting with a digit get an underscore
// prefix, and empty names become "unnamed".
func SanitizeName(name string) string {
	var b strings.Builder

	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == ' ' || r == '-':
			b.WriteRune('_')
		case r == '_':
			b.WriteRune('_')
		}
		// all other characters are dropped
	}

	result := b.String()

	if result == "" {
		return "unnamed"
	}

	// Terraform identifiers cannot start with a digit
	if unicode.IsDigit(rune(result[0])) {
		result = "_" + result
	}

	return result
}

// GenerateProvider returns the content of provider.tf, pinning the
// filipowm/unifi provider to ~> 1.0.0 and declaring variables for
// the API URL and key.
func GenerateProvider() string {
	return `terraform {
  required_providers {
    unifi = {
      source  = "filipowm/unifi"
      version = "~> 1.0.0"
    }
  }
}

variable "unifi_api_url" {
  description = "URL of the UniFi controller"
  type        = string
}

variable "unifi_api_key" {
  description = "API key for the UniFi controller"
  type        = string
  sensitive   = true
}

provider "unifi" {
  api_url  = var.unifi_api_url
  api_key  = var.unifi_api_key
  insecure = true
}
`
}

// importBlock returns a single Terraform import block for the given
// resource type, sanitized name, and ID.
func importBlock(resourceType, sanitizedName, id string) string {
	return fmt.Sprintf(`import {
  to = %s.%s
  id = "%s"
}
`, resourceType, sanitizedName, id)
}

// GenerateImports returns the content of imports.tf for the given
// inventory. One import block is generated per non-device resource.
// Devices are excluded because they have no Terraform resource type.
func GenerateImports(inv *models.Inventory) string {
	var b strings.Builder

	for _, n := range inv.Networks {
		b.WriteString(importBlock("unifi_network", SanitizeName(n.Name), n.ID))
		b.WriteString("\n")
	}
	for _, r := range inv.FirewallRules {
		b.WriteString(importBlock("unifi_firewall_rule", SanitizeName(r.Name), r.ID))
		b.WriteString("\n")
	}
	for _, g := range inv.FirewallGroups {
		b.WriteString(importBlock("unifi_firewall_group", SanitizeName(g.Name), g.ID))
		b.WriteString("\n")
	}
	for _, w := range inv.WLANs {
		b.WriteString(importBlock("unifi_wlan", SanitizeName(w.Name), w.ID))
		b.WriteString("\n")
	}
	for _, p := range inv.PortForwards {
		b.WriteString(importBlock("unifi_port_forward", SanitizeName(p.Name), p.ID))
		b.WriteString("\n")
	}
	for _, p := range inv.PortProfiles {
		b.WriteString(importBlock("unifi_port_profile", SanitizeName(p.Name), p.ID))
		b.WriteString("\n")
	}
	for _, s := range inv.StaticRoutes {
		b.WriteString(importBlock("unifi_static_route", SanitizeName(s.Name), s.ID))
		b.WriteString("\n")
	}

	// Trim trailing newline if we wrote anything
	result := b.String()
	if len(result) > 0 {
		result = strings.TrimSuffix(result, "\n")
	}

	return result
}

// stubBlock returns a single Terraform resource stub for a non-WLAN
// resource.
func stubBlock(resourceType, sanitizedName string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
  # Stub — will be replaced by terraform plan -generate-config-out
}
`, resourceType, sanitizedName)
}

// wlanStubBlock returns a Terraform resource stub for a WLAN resource,
// including a passphrase placeholder instead of the real value.
func wlanStubBlock(sanitizedName string) string {
	return fmt.Sprintf(`resource "unifi_wlan" "%s" {
  # Stub — will be replaced by terraform plan -generate-config-out
  x_passphrase = "REPLACE_WITH_ACTUAL_PASSPHRASE"  # REPLACE with real passphrase
}
`, sanitizedName)
}

// GenerateStubs returns the content of stubs.tf for the given
// inventory. Minimal HCL resource skeletons are generated for each
// non-device resource. WLAN stubs include a passphrase placeholder
// to ensure no real passphrase values appear in generated output.
func GenerateStubs(inv *models.Inventory) string {
	var b strings.Builder

	for _, n := range inv.Networks {
		b.WriteString(stubBlock("unifi_network", SanitizeName(n.Name)))
		b.WriteString("\n")
	}
	for _, r := range inv.FirewallRules {
		b.WriteString(stubBlock("unifi_firewall_rule", SanitizeName(r.Name)))
		b.WriteString("\n")
	}
	for _, g := range inv.FirewallGroups {
		b.WriteString(stubBlock("unifi_firewall_group", SanitizeName(g.Name)))
		b.WriteString("\n")
	}
	for _, w := range inv.WLANs {
		b.WriteString(wlanStubBlock(SanitizeName(w.Name)))
		b.WriteString("\n")
	}
	for _, p := range inv.PortForwards {
		b.WriteString(stubBlock("unifi_port_forward", SanitizeName(p.Name)))
		b.WriteString("\n")
	}
	for _, p := range inv.PortProfiles {
		b.WriteString(stubBlock("unifi_port_profile", SanitizeName(p.Name)))
		b.WriteString("\n")
	}
	for _, s := range inv.StaticRoutes {
		b.WriteString(stubBlock("unifi_static_route", SanitizeName(s.Name)))
		b.WriteString("\n")
	}

	// Trim trailing newline if we wrote anything
	result := b.String()
	if len(result) > 0 {
		result = strings.TrimSuffix(result, "\n")
	}

	return result
}
