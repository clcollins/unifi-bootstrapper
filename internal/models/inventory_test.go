package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventory_RoundTrip(t *testing.T) {
	exportTime := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)

	original := Inventory{
		Networks: []Network{
			{
				ID:      "net1",
				Name:    "LAN",
				Purpose: "corporate",
				Subnet:  "10.0.0.0/24",
				SiteID:  "site1",
			},
		},
		FirewallRules: []FirewallRule{
			{
				ID:        "fw1",
				Name:      "Allow All",
				Action:    "accept",
				Enabled:   true,
				RuleIndex: FlexInt(1000),
				Ruleset:   "LAN_IN",
				SiteID:    "site1",
			},
		},
		FirewallGroups: []FirewallGroup{
			{
				ID:           "fg1",
				Name:         "Addresses",
				GroupType:    "address-group",
				GroupMembers: []string{"10.0.0.0/8"},
				SiteID:       "site1",
			},
		},
		WLANs: []WLAN{
			{
				ID:          "wlan1",
				Name:        "WiFi",
				Enabled:     true,
				Security:    "wpapsk",
				XPassphrase: "fake-passphrase",
				SiteID:      "site1",
			},
		},
		PortForwards: []PortForward{
			{
				ID:      "pf1",
				Name:    "HTTP",
				Enabled: true,
				DstPort: "80",
				Fwd:     "10.0.0.100",
				FwdPort: "80",
				Proto:   "tcp",
				SiteID:  "site1",
			},
		},
		PortProfiles: []PortProfile{
			{
				ID:                   "pp1",
				Name:                 "Default",
				TaggedNetworkconfIDs: []string{},
				SiteID:               "site1",
			},
		},
		StaticRoutes: []StaticRoute{
			{
				ID:                 "sr1",
				Name:               "Route",
				Enabled:            true,
				Type:               "nexthop-route",
				StaticRouteNetwork: "172.16.0.0/24",
				SiteID:             "site1",
			},
		},
		Devices: []Device{
			{
				ID:      "dev1",
				Name:    "Gateway",
				Model:   "UDM",
				Type:    "udm",
				Mac:     "aa:bb:cc:dd:ee:ff",
				IP:      "10.0.0.1",
				Adopted: true,
				State:   1,
				SiteID:  "site1",
			},
		},
		ExportedAt: exportTime,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Inventory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestInventory_EmptySlices(t *testing.T) {
	inv := Inventory{
		Networks:       []Network{},
		FirewallRules:  []FirewallRule{},
		FirewallGroups: []FirewallGroup{},
		WLANs:          []WLAN{},
		PortForwards:   []PortForward{},
		PortProfiles:   []PortProfile{},
		StaticRoutes:   []StaticRoute{},
		Devices:        []Device{},
		ExportedAt:     time.Now().UTC(),
	}

	data, err := json.Marshal(inv)
	require.NoError(t, err)

	var decoded Inventory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.Networks)
	assert.Empty(t, decoded.FirewallRules)
	assert.Empty(t, decoded.FirewallGroups)
	assert.Empty(t, decoded.WLANs)
	assert.Empty(t, decoded.PortForwards)
	assert.Empty(t, decoded.PortProfiles)
	assert.Empty(t, decoded.StaticRoutes)
	assert.Empty(t, decoded.Devices)
}

func TestInventory_ExportedAtTimestamp(t *testing.T) {
	exportTime := time.Date(2026, 5, 25, 15, 30, 0, 0, time.UTC)
	inv := Inventory{
		ExportedAt: exportTime,
	}

	data, err := json.Marshal(inv)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	// Verify the timestamp is in RFC3339 format
	assert.Contains(t, raw["exported_at"].(string), "2026-05-25T15:30:00")
}

func TestInventory_ContainsAllResourceTypes(t *testing.T) {
	// Verify the Inventory struct has all expected fields by
	// checking that JSON keys are present
	inv := Inventory{
		Networks:       []Network{{ID: "n1", SiteID: "s"}},
		FirewallRules:  []FirewallRule{{ID: "fr1", SiteID: "s"}},
		FirewallGroups: []FirewallGroup{{ID: "fg1", SiteID: "s"}},
		WLANs:          []WLAN{{ID: "w1", SiteID: "s"}},
		PortForwards:   []PortForward{{ID: "pf1", SiteID: "s", DstPort: "80", Fwd: "10.0.0.1", FwdPort: "80", Proto: "tcp"}},
		PortProfiles:   []PortProfile{{ID: "pp1", SiteID: "s"}},
		StaticRoutes:   []StaticRoute{{ID: "sr1", SiteID: "s", StaticRouteNetwork: "10.0.0.0/8"}},
		Devices:        []Device{{ID: "d1", SiteID: "s", Mac: "aa:bb:cc:dd:ee:ff", IP: "10.0.0.1"}},
		ExportedAt:     time.Now().UTC(),
	}

	data, err := json.Marshal(inv)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	expectedKeys := []string{
		"networks",
		"firewall_rules",
		"firewall_groups",
		"wlans",
		"port_forwards",
		"port_profiles",
		"static_routes",
		"devices",
		"exported_at",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, raw, key, "Inventory JSON missing key: %s", key)
	}
}
