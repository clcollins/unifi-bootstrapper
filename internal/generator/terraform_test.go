package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "spaces become underscores",
			input:    "My LAN Network",
			expected: "My_LAN_Network",
		},
		{
			name:     "special characters removed",
			input:    "net@work#1!",
			expected: "network1",
		},
		{
			name:     "leading digit gets underscore prefix",
			input:    "5GHz WiFi",
			expected: "_5GHz_WiFi",
		},
		{
			name:     "empty name becomes unnamed",
			input:    "",
			expected: "unnamed",
		},
		{
			name:     "already valid name unchanged",
			input:    "my_lan",
			expected: "my_lan",
		},
		{
			name:     "hyphens become underscores",
			input:    "guest-network",
			expected: "guest_network",
		},
		{
			name:     "mixed special characters",
			input:    "VLAN (100) - Management",
			expected: "VLAN_100___Management",
		},
		{
			name:     "only special characters becomes unnamed",
			input:    "@#$%",
			expected: "unnamed",
		},
		{
			name:     "leading digit after sanitization",
			input:    "100_servers",
			expected: "_100_servers",
		},
		{
			name:     "underscores preserved",
			input:    "my_network_name",
			expected: "my_network_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeName_IsPure(t *testing.T) {
	// Calling with the same input must always return the same output
	input := "My Network (5GHz)"
	first := SanitizeName(input)
	second := SanitizeName(input)
	assert.Equal(t, first, second, "SanitizeName must be pure — same input must produce same output")
}

func TestGenerateProvider(t *testing.T) {
	result := GenerateProvider()

	// Verify required_providers block
	assert.Contains(t, result, `source  = "filipowm/unifi"`)
	assert.Contains(t, result, `version = "~> 1.0.0"`)

	// Verify variable blocks
	assert.Contains(t, result, `variable "unifi_api_url"`)
	assert.Contains(t, result, `variable "unifi_api_key"`)
	assert.Contains(t, result, `sensitive   = true`)

	// Verify provider block
	assert.Contains(t, result, `provider "unifi"`)
	assert.Contains(t, result, `api_url  = var.unifi_api_url`)
	assert.Contains(t, result, `api_key  = var.unifi_api_key`)
	assert.Contains(t, result, `insecure = true`)
}

func TestGenerateProvider_IsPure(t *testing.T) {
	first := GenerateProvider()
	second := GenerateProvider()
	assert.Equal(t, first, second, "GenerateProvider must be pure — same output every call")
}

func TestGenerateProvider_GoldenFile(t *testing.T) {
	expected := `terraform {
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
	result := GenerateProvider()
	assert.Equal(t, expected, result)
}

func buildTestInventory() *models.Inventory {
	return &models.Inventory{
		Networks: []models.Network{
			{
				ID:      "aaa111000000000000000001",
				Name:    "Default LAN",
				Purpose: "corporate",
				Subnet:  "10.0.1.0/24",
				SiteID:  "site1",
			},
			{
				ID:          "aaa111000000000000000002",
				Name:        "Guest Network",
				Purpose:     "guest",
				Subnet:      "10.0.50.0/24",
				VLANID:      50,
				VLANEnabled: true,
				SiteID:      "site1",
			},
		},
		FirewallRules: []models.FirewallRule{
			{
				ID:        "bbb222000000000000000001",
				Name:      "Allow LAN to WAN",
				Action:    "accept",
				Enabled:   true,
				RuleIndex: models.FlexInt(2000),
				Ruleset:   "LAN_IN",
				SiteID:    "site1",
			},
		},
		FirewallGroups: []models.FirewallGroup{
			{
				ID:           "ccc333000000000000000001",
				Name:         "RFC1918 Addresses",
				GroupType:    "address-group",
				GroupMembers: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
				SiteID:       "site1",
			},
		},
		WLANs: []models.WLAN{
			{
				ID:          "ddd444000000000000000001",
				Name:        "HomeNet 5G",
				Enabled:     true,
				Security:    "wpapsk",
				XPassphrase: "super-secret-passphrase",
				SiteID:      "site1",
			},
		},
		PortForwards: []models.PortForward{
			{
				ID:      "eee555000000000000000001",
				Name:    "Web Server",
				Enabled: true,
				DstPort: "443",
				Fwd:     "10.0.1.100",
				FwdPort: "443",
				Proto:   "tcp",
				SiteID:  "site1",
			},
		},
		PortProfiles: []models.PortProfile{
			{
				ID:                   "fff666000000000000000001",
				Name:                 "All",
				TaggedNetworkconfIDs: []string{},
				SiteID:               "site1",
			},
		},
		StaticRoutes: []models.StaticRoute{
			{
				ID:                 "ggg777000000000000000001",
				Name:               "VPN Route",
				Enabled:            true,
				Type:               "nexthop-route",
				StaticRouteNetwork: "172.16.0.0/24",
				GatewayIP:          "10.0.1.1",
				SiteID:             "site1",
			},
		},
		Devices: []models.Device{
			{
				ID:      "hhh888000000000000000001",
				Name:    "UDM Pro",
				Model:   "UDM",
				Type:    "udm",
				Mac:     "aa:bb:cc:dd:ee:ff",
				IP:      "10.0.1.1",
				Adopted: true,
				State:   1,
				SiteID:  "site1",
			},
		},
	}
}

func TestGenerateImports(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateImports(inv)

	// Verify one import block per non-device resource
	// 2 networks + 1 firewall rule + 1 firewall group + 1 WLAN +
	// 1 port forward + 1 port profile + 1 static route = 8
	importCount := strings.Count(result, "import {")
	assert.Equal(t, 8, importCount, "expected 8 import blocks (one per non-device resource)")

	// Verify resource type mapping
	assert.Contains(t, result, `to = unifi_network.Default_LAN`)
	assert.Contains(t, result, `to = unifi_network.Guest_Network`)
	assert.Contains(t, result, `to = unifi_firewall_rule.Allow_LAN_to_WAN`)
	assert.Contains(t, result, `to = unifi_firewall_group.RFC1918_Addresses`)
	assert.Contains(t, result, `to = unifi_wlan.HomeNet_5G`)
	assert.Contains(t, result, `to = unifi_port_forward.Web_Server`)
	assert.Contains(t, result, `to = unifi_port_profile.All`)
	assert.Contains(t, result, `to = unifi_static_route.VPN_Route`)

	// Verify IDs are present
	assert.Contains(t, result, `id = "aaa111000000000000000001"`)
	assert.Contains(t, result, `id = "aaa111000000000000000002"`)
	assert.Contains(t, result, `id = "bbb222000000000000000001"`)
	assert.Contains(t, result, `id = "ccc333000000000000000001"`)
	assert.Contains(t, result, `id = "ddd444000000000000000001"`)
	assert.Contains(t, result, `id = "eee555000000000000000001"`)
	assert.Contains(t, result, `id = "fff666000000000000000001"`)
	assert.Contains(t, result, `id = "ggg777000000000000000001"`)

	// Verify devices are excluded
	assert.NotContains(t, result, "hhh888000000000000000001")
	assert.NotContains(t, result, "UDM_Pro")
}

func TestGenerateImports_IsPure(t *testing.T) {
	inv := buildTestInventory()
	first := GenerateImports(inv)
	second := GenerateImports(inv)
	assert.Equal(t, first, second, "GenerateImports must be pure — same input must produce same output")
}

func TestGenerateImports_EmptyInventory(t *testing.T) {
	inv := &models.Inventory{
		Networks:       []models.Network{},
		FirewallRules:  []models.FirewallRule{},
		FirewallGroups: []models.FirewallGroup{},
		WLANs:          []models.WLAN{},
		PortForwards:   []models.PortForward{},
		PortProfiles:   []models.PortProfile{},
		StaticRoutes:   []models.StaticRoute{},
		Devices:        []models.Device{},
	}
	result := GenerateImports(inv)
	assert.Empty(t, result, "empty inventory should produce empty imports")
}

func TestGenerateImports_NilSlices(t *testing.T) {
	inv := &models.Inventory{}
	result := GenerateImports(inv)
	assert.Empty(t, result, "inventory with nil slices should produce empty imports")
}

func TestGenerateImports_GoldenFile(t *testing.T) {
	inv := &models.Inventory{
		Networks: []models.Network{
			{
				ID:     "net001",
				Name:   "My LAN",
				SiteID: "site1",
			},
		},
		FirewallRules: []models.FirewallRule{
			{
				ID:     "fw001",
				Name:   "Block Bad",
				SiteID: "site1",
			},
		},
	}

	expected := `import {
  to = unifi_network.My_LAN
  id = "net001"
}

import {
  to = unifi_firewall_rule.Block_Bad
  id = "fw001"
}
`
	result := GenerateImports(inv)
	assert.Equal(t, expected, result)
}

func TestGenerateImports_NameSanitization(t *testing.T) {
	inv := &models.Inventory{
		Networks: []models.Network{
			{
				ID:     "net001",
				Name:   "5GHz Band",
				SiteID: "site1",
			},
		},
	}
	result := GenerateImports(inv)
	// Leading digit should get underscore prefix
	assert.Contains(t, result, `to = unifi_network._5GHz_Band`)
}

func TestGenerateStubs(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateStubs(inv)

	// Verify one resource block per non-device resource
	resourceCount := strings.Count(result, "resource ")
	assert.Equal(t, 8, resourceCount, "expected 8 resource stubs (one per non-device resource)")

	// Verify resource type mapping
	assert.Contains(t, result, `resource "unifi_network" "Default_LAN"`)
	assert.Contains(t, result, `resource "unifi_network" "Guest_Network"`)
	assert.Contains(t, result, `resource "unifi_firewall_rule" "Allow_LAN_to_WAN"`)
	assert.Contains(t, result, `resource "unifi_firewall_group" "RFC1918_Addresses"`)
	assert.Contains(t, result, `resource "unifi_wlan" "HomeNet_5G"`)
	assert.Contains(t, result, `resource "unifi_port_forward" "Web_Server"`)
	assert.Contains(t, result, `resource "unifi_port_profile" "All"`)
	assert.Contains(t, result, `resource "unifi_static_route" "VPN_Route"`)

	// Verify stub comment present
	assert.Contains(t, result, "# Stub")

	// Verify devices are excluded
	assert.NotContains(t, result, `"UDM_Pro"`)
	assert.NotContains(t, result, "unifi_device")
}

func TestGenerateStubs_WLANContainsPassphrasePlaceholder(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateStubs(inv)

	// WLAN stubs must contain placeholder, not real passphrase
	assert.Contains(t, result, `x_passphrase = "REPLACE_WITH_ACTUAL_PASSPHRASE"`)
	assert.Contains(t, result, "# REPLACE with real passphrase")
	assert.NotContains(t, result, "super-secret-passphrase",
		"real passphrase must never appear in stubs")
}

func TestGenerateStubs_IsPure(t *testing.T) {
	inv := buildTestInventory()
	first := GenerateStubs(inv)
	second := GenerateStubs(inv)
	assert.Equal(t, first, second, "GenerateStubs must be pure — same input must produce same output")
}

func TestGenerateStubs_EmptyInventory(t *testing.T) {
	inv := &models.Inventory{
		Networks:       []models.Network{},
		FirewallRules:  []models.FirewallRule{},
		FirewallGroups: []models.FirewallGroup{},
		WLANs:          []models.WLAN{},
		PortForwards:   []models.PortForward{},
		PortProfiles:   []models.PortProfile{},
		StaticRoutes:   []models.StaticRoute{},
		Devices:        []models.Device{},
	}
	result := GenerateStubs(inv)
	assert.Empty(t, result, "empty inventory should produce empty stubs")
}

func TestGenerateStubs_NilSlices(t *testing.T) {
	inv := &models.Inventory{}
	result := GenerateStubs(inv)
	assert.Empty(t, result, "inventory with nil slices should produce empty stubs")
}

func TestGenerateStubs_GoldenFile_NonWLAN(t *testing.T) {
	inv := &models.Inventory{
		Networks: []models.Network{
			{
				ID:     "net001",
				Name:   "My LAN",
				SiteID: "site1",
			},
		},
	}

	expected := `resource "unifi_network" "My_LAN" {
  # Stub — will be replaced by terraform plan -generate-config-out
}
`
	result := GenerateStubs(inv)
	assert.Equal(t, expected, result)
}

func TestGenerateStubs_GoldenFile_WLAN(t *testing.T) {
	inv := &models.Inventory{
		WLANs: []models.WLAN{
			{
				ID:          "wlan001",
				Name:        "HomeNet 5G",
				XPassphrase: "do-not-show-this",
				SiteID:      "site1",
			},
		},
	}

	expected := `resource "unifi_wlan" "HomeNet_5G" {
  # Stub — will be replaced by terraform plan -generate-config-out
  x_passphrase = "REPLACE_WITH_ACTUAL_PASSPHRASE"  # REPLACE with real passphrase
}
`
	result := GenerateStubs(inv)
	assert.Equal(t, expected, result)
}

func TestGenerateStubs_MultipleWLANs_NoneLeakPassphrase(t *testing.T) {
	inv := &models.Inventory{
		WLANs: []models.WLAN{
			{
				ID:          "wlan001",
				Name:        "Net A",
				XPassphrase: "secret-a",
				SiteID:      "site1",
			},
			{
				ID:          "wlan002",
				Name:        "Net B",
				XPassphrase: "secret-b",
				SiteID:      "site1",
			},
		},
	}
	result := GenerateStubs(inv)

	assert.NotContains(t, result, "secret-a")
	assert.NotContains(t, result, "secret-b")

	placeholderCount := strings.Count(result, `x_passphrase = "REPLACE_WITH_ACTUAL_PASSPHRASE"`)
	assert.Equal(t, 2, placeholderCount, "each WLAN stub must have a passphrase placeholder")
}

func TestGenerateImports_DevicesOnlyInventory(t *testing.T) {
	inv := &models.Inventory{
		Devices: []models.Device{
			{
				ID:     "dev001",
				Name:   "Switch",
				Model:  "USW",
				Type:   "usw",
				Mac:    "aa:bb:cc:dd:ee:00",
				IP:     "10.0.0.2",
				SiteID: "site1",
			},
		},
	}
	result := GenerateImports(inv)
	assert.Empty(t, result, "devices-only inventory should produce empty imports")
}

func TestGenerateStubs_DevicesOnlyInventory(t *testing.T) {
	inv := &models.Inventory{
		Devices: []models.Device{
			{
				ID:     "dev001",
				Name:   "Switch",
				Model:  "USW",
				Type:   "usw",
				Mac:    "aa:bb:cc:dd:ee:00",
				IP:     "10.0.0.2",
				SiteID: "site1",
			},
		},
	}
	result := GenerateStubs(inv)
	assert.Empty(t, result, "devices-only inventory should produce empty stubs")
}

func TestGenerateImports_AllResourceTypes(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateImports(inv)

	// Verify all resource types are represented
	expectedTypes := []string{
		"unifi_network",
		"unifi_firewall_rule",
		"unifi_firewall_group",
		"unifi_wlan",
		"unifi_port_forward",
		"unifi_port_profile",
		"unifi_static_route",
	}
	for _, resType := range expectedTypes {
		assert.Contains(t, result, resType,
			"imports should contain resource type %s", resType)
	}
}

func TestGenerateStubs_AllResourceTypes(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateStubs(inv)

	expectedTypes := []string{
		"unifi_network",
		"unifi_firewall_rule",
		"unifi_firewall_group",
		"unifi_wlan",
		"unifi_port_forward",
		"unifi_port_profile",
		"unifi_static_route",
	}
	for _, resType := range expectedTypes {
		assert.Contains(t, result, resType,
			"stubs should contain resource type %s", resType)
	}
}

func TestGenerateImports_FullGoldenFile(t *testing.T) {
	inv := buildTestInventory()
	result := GenerateImports(inv)

	// Count total blocks — should match 8 non-device resources
	require.Equal(t, 8, strings.Count(result, "import {"))

	// Verify each block has matching to and id lines
	blocks := strings.Split(result, "import {")
	// First element is empty (before first "import {")
	for i := 1; i < len(blocks); i++ {
		block := blocks[i]
		assert.Contains(t, block, "to = ", "block %d missing 'to' line", i)
		assert.Contains(t, block, "id = ", "block %d missing 'id' line", i)
		assert.Contains(t, block, "}", "block %d missing closing brace", i)
	}
}
