package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/networks.json")
	require.NoError(t, err)

	var resp APIResponse[Network]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 3)

	// Verify the default LAN
	lan := resp.Data[0]
	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", lan.ID)
	assert.Equal(t, "Default LAN", lan.Name)
	assert.Equal(t, "corporate", lan.Purpose)
	assert.Equal(t, "10.0.1.0/24", lan.Subnet)
	assert.False(t, lan.VLANEnabled)
	assert.True(t, lan.DHCPEnabled)
	assert.Equal(t, "10.0.1.100", lan.DHCPStart)
	assert.Equal(t, "10.0.1.254", lan.DHCPStop)
	assert.True(t, lan.DHCPDNSEnabled)
	assert.Equal(t, "localdomain", lan.DomainName)
	assert.True(t, lan.IGMPSnooping)
	assert.Equal(t, "LAN", lan.NetworkGroup)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", lan.SiteID)

	// Verify the guest VLAN
	guest := resp.Data[1]
	assert.Equal(t, "Guest Network", guest.Name)
	assert.Equal(t, "guest", guest.Purpose)
	assert.Equal(t, 50, guest.VLANID)
	assert.True(t, guest.VLANEnabled)
	assert.False(t, guest.DHCPDNSEnabled)

	// Verify the management VLAN
	mgmt := resp.Data[2]
	assert.Equal(t, "Management VLAN", mgmt.Name)
	assert.Equal(t, 99, mgmt.VLANID)
	assert.True(t, mgmt.VLANEnabled)
	assert.Equal(t, "mgmt.localdomain", mgmt.DomainName)
}

func TestNetwork_RoundTrip(t *testing.T) {
	original := Network{
		ID:             "abc123",
		Name:           "Test Net",
		Purpose:        "corporate",
		Subnet:         "10.10.0.0/24",
		VLANID:         100,
		VLANEnabled:    true,
		DHCPEnabled:    true,
		DHCPStart:      "10.10.0.100",
		DHCPStop:       "10.10.0.200",
		DHCPDNSEnabled: true,
		DomainName:     "test.local",
		IGMPSnooping:   false,
		NetworkGroup:   "LAN",
		SiteID:         "site123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Network
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestNetwork_OmitEmptyFields(t *testing.T) {
	net := Network{
		ID:      "abc123",
		Name:    "Minimal",
		Purpose: "corporate",
		SiteID:  "site123",
	}

	data, err := json.Marshal(net)
	require.NoError(t, err)

	// Fields with omitempty should be absent when zero-valued
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "subnet")
	assert.NotContains(t, raw, "vlan")
	assert.NotContains(t, raw, "dhcpd_start")
	assert.NotContains(t, raw, "dhcpd_stop")
	assert.NotContains(t, raw, "domain_name")
	assert.NotContains(t, raw, "networkgroup")

	// Non-omitempty fields should be present even when zero
	assert.Contains(t, raw, "_id")
	assert.Contains(t, raw, "name")
	assert.Contains(t, raw, "purpose")
	assert.Contains(t, raw, "vlan_enabled")
	assert.Contains(t, raw, "dhcp_enabled")
	assert.Contains(t, raw, "dhcpd_dns_enabled")
	assert.Contains(t, raw, "igmp_snooping")
	assert.Contains(t, raw, "site_id")
}
