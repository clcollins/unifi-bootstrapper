package renderer

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// knownPassphrases are fake values used to verify redaction in rendered output.
var knownPassphrases = []string{
	"test-passphrase-do-not-use",
	"guest-fake-password-not-real",
}

// testInventory returns a populated Inventory with all resource types for testing.
func testInventory() *models.Inventory {
	return &models.Inventory{
		ExportedAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
		Networks: []models.Network{
			{
				ID:          "net-001",
				Name:        "LAN",
				Purpose:     "corporate",
				Subnet:      "192.168.1.0/24",
				VLANID:      1,
				VLANEnabled: false,
				DHCPEnabled: true,
				SiteID:      "site-001",
			},
			{
				ID:          "net-002",
				Name:        "Guest",
				Purpose:     "guest",
				Subnet:      "192.168.50.0/24",
				VLANID:      50,
				VLANEnabled: true,
				DHCPEnabled: true,
				SiteID:      "site-001",
			},
		},
		FirewallRules: []models.FirewallRule{
			{
				ID:       "fw-001",
				Name:     "Block IoT to LAN",
				Action:   "drop",
				Enabled:  true,
				Ruleset:  "LAN_IN",
				Protocol: "all",
				SiteID:   "site-001",
			},
		},
		FirewallGroups: []models.FirewallGroup{
			{
				ID:           "fg-001",
				Name:         "RFC1918",
				GroupType:    "address-group",
				GroupMembers: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
				SiteID:       "site-001",
			},
		},
		WLANs: []models.WLAN{
			{
				ID:          "wlan-001",
				Name:        "HomeNet-5G",
				Enabled:     true,
				Security:    "wpapsk",
				WPAMode:     "wpa2",
				XPassphrase: knownPassphrases[0],
				VLANID:      1,
				VLANEnabled: false,
				WlanBand:    "both",
				SiteID:      "site-001",
			},
			{
				ID:          "wlan-002",
				Name:        "GuestWiFi",
				Enabled:     true,
				Security:    "wpapsk",
				WPAMode:     "wpa2",
				XPassphrase: knownPassphrases[1],
				VLANID:      50,
				VLANEnabled: true,
				IsGuest:     true,
				WlanBand:    "both",
				SiteID:      "site-001",
			},
		},
		PortForwards: []models.PortForward{
			{
				ID:      "pf-001",
				Name:    "HTTP Forward",
				Enabled: true,
				DstPort: "80",
				Fwd:     "192.168.1.100",
				FwdPort: "8080",
				Proto:   "tcp",
				SiteID:  "site-001",
			},
		},
		PortProfiles: []models.PortProfile{
			{
				ID:      "pp-001",
				Name:    "All",
				Forward: "all",
				SiteID:  "site-001",
			},
		},
		StaticRoutes: []models.StaticRoute{
			{
				ID:                 "sr-001",
				Name:               "VPN Route",
				Enabled:            true,
				Type:               "nexthop-route",
				StaticRouteNetwork: "10.0.0.0/8",
				GatewayIP:          "192.168.1.1",
				SiteID:             "site-001",
			},
		},
		Devices: []models.Device{
			{
				ID:      "dev-001",
				Name:    "UDM-Pro",
				Model:   "UDM-Pro",
				Type:    "udm",
				Mac:     "aa:bb:cc:dd:ee:ff",
				IP:      "192.168.1.1",
				Version: "3.0.20",
				Adopted: true,
				State:   1,
				SiteID:  "site-001",
			},
		},
	}
}

func TestRenderJSON_HappyPath(t *testing.T) {
	inv := testInventory()

	output, err := RenderJSON(inv)
	require.NoError(t, err)
	assert.True(t, json.Valid(output), "output must be valid JSON")

	// Verify it's indented (contains newlines and spaces)
	assert.Contains(t, string(output), "\n")
	assert.Contains(t, string(output), "  ")
}

func TestRenderJSON_PassphraseRedaction(t *testing.T) {
	inv := testInventory()

	output, err := RenderJSON(inv)
	require.NoError(t, err)

	// The output MUST contain REDACTED where passphrases were
	assert.Contains(t, string(output), `"REDACTED"`,
		"output must contain REDACTED for passphrase fields")

	// The output MUST NOT contain ANY of the known passphrase values
	for _, passphrase := range knownPassphrases {
		assert.False(t, bytes.Contains(output, []byte(passphrase)),
			"output must NOT contain known passphrase %q", passphrase)
	}
}

func TestRenderJSON_EmptyInventory(t *testing.T) {
	inv := &models.Inventory{
		ExportedAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
	}

	output, err := RenderJSON(inv)
	require.NoError(t, err)
	assert.True(t, json.Valid(output), "output must be valid JSON")

	// Unmarshal and verify empty slices
	var decoded models.Inventory
	err = json.Unmarshal(output, &decoded)
	require.NoError(t, err)
	assert.Empty(t, decoded.Networks)
	assert.Empty(t, decoded.WLANs)
	assert.Empty(t, decoded.Devices)
}

func TestRenderJSON_RoundTrip(t *testing.T) {
	inv := testInventory()

	output, err := RenderJSON(inv)
	require.NoError(t, err)

	var decoded models.Inventory
	err = json.Unmarshal(output, &decoded)
	require.NoError(t, err)

	// Structural integrity: counts match
	assert.Len(t, decoded.Networks, len(inv.Networks))
	assert.Len(t, decoded.FirewallRules, len(inv.FirewallRules))
	assert.Len(t, decoded.FirewallGroups, len(inv.FirewallGroups))
	assert.Len(t, decoded.WLANs, len(inv.WLANs))
	assert.Len(t, decoded.PortForwards, len(inv.PortForwards))
	assert.Len(t, decoded.PortProfiles, len(inv.PortProfiles))
	assert.Len(t, decoded.StaticRoutes, len(inv.StaticRoutes))
	assert.Len(t, decoded.Devices, len(inv.Devices))

	// Names preserved
	assert.Equal(t, "LAN", decoded.Networks[0].Name)
	assert.Equal(t, "HomeNet-5G", decoded.WLANs[0].Name)
	assert.Equal(t, "UDM-Pro", decoded.Devices[0].Name)

	// Passphrases MUST be REDACTED in the decoded output
	for _, wlan := range decoded.WLANs {
		assert.Equal(t, "REDACTED", wlan.XPassphrase,
			"passphrase for WLAN %q must be REDACTED after round-trip", wlan.Name)
	}
}

func TestRenderJSON_DoesNotMutateInput(t *testing.T) {
	inv := testInventory()
	originalPassphrase := inv.WLANs[0].XPassphrase

	_, err := RenderJSON(inv)
	require.NoError(t, err)

	// The original inventory must NOT be modified
	assert.Equal(t, originalPassphrase, inv.WLANs[0].XPassphrase,
		"RenderJSON must not mutate the input inventory")
}
