package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWLAN_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/wlans.json")
	require.NoError(t, err)

	var resp APIResponse[WLAN]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// Primary WLAN
	primary := resp.Data[0]
	assert.Equal(t, "f6a7b8c9d0e1f2a3b4c5d6e7", primary.ID)
	assert.Equal(t, "HomeNet-5G", primary.Name)
	assert.True(t, primary.Enabled)
	assert.Equal(t, "wpapsk", primary.Security)
	assert.Equal(t, "wpa2", primary.WPAMode)
	assert.Equal(t, "test-passphrase-do-not-use", primary.XPassphrase)
	assert.False(t, primary.VLANEnabled)
	assert.False(t, primary.IsGuest)
	assert.False(t, primary.HideSsid)
	assert.Equal(t, "both", primary.WlanBand)
	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", primary.NetworkID)
	assert.Equal(t, "a1a2a3a4a5a6a7a8a9a0a1a2", primary.UserGroupID)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", primary.SiteID)

	// Guest WLAN
	guest := resp.Data[1]
	assert.Equal(t, "GuestWiFi", guest.Name)
	assert.Equal(t, "guest-fake-password-not-real", guest.XPassphrase)
	assert.Equal(t, 50, guest.VLANID)
	assert.True(t, guest.VLANEnabled)
	assert.True(t, guest.IsGuest)
}

func TestWLAN_XPassphrasePresent(t *testing.T) {
	// Verify that x_passphrase is present in the model and correctly
	// round-trips through JSON
	wlan := WLAN{
		ID:          "test123",
		Name:        "TestWLAN",
		Enabled:     true,
		Security:    "wpapsk",
		WPAMode:     "wpa2",
		XPassphrase: "fake-secret-value",
		SiteID:      "site123",
	}

	data, err := json.Marshal(wlan)
	require.NoError(t, err)

	// Verify the field is in the JSON output
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	assert.Equal(t, "fake-secret-value", raw["x_passphrase"])

	// Round-trip back to struct
	var decoded WLAN
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "fake-secret-value", decoded.XPassphrase)
}

func TestWLAN_RoundTrip(t *testing.T) {
	original := WLAN{
		ID:          "abc123",
		Name:        "Test WLAN",
		Enabled:     true,
		Security:    "wpapsk",
		WPAMode:     "wpa2",
		XPassphrase: "not-a-real-passphrase",
		VLANID:      10,
		VLANEnabled: true,
		IsGuest:     false,
		HideSsid:    true,
		WlanBand:    "5g",
		NetworkID:   "net123",
		UserGroupID: "ug123",
		SiteID:      "site123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded WLAN
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestWLAN_OmitEmptyFields(t *testing.T) {
	wlan := WLAN{
		ID:       "abc123",
		Name:     "Minimal",
		Enabled:  true,
		Security: "open",
		SiteID:   "site123",
	}

	data, err := json.Marshal(wlan)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	// omitempty fields absent when zero
	assert.NotContains(t, raw, "wpa_mode")
	assert.NotContains(t, raw, "x_passphrase")
	assert.NotContains(t, raw, "vlan")
	assert.NotContains(t, raw, "wlan_band")
	assert.NotContains(t, raw, "networkconf_id")
	assert.NotContains(t, raw, "usergroup_id")

	// Non-omitempty fields present
	assert.Contains(t, raw, "_id")
	assert.Contains(t, raw, "name")
	assert.Contains(t, raw, "enabled")
	assert.Contains(t, raw, "security")
	assert.Contains(t, raw, "vlan_enabled")
	assert.Contains(t, raw, "is_guest")
	assert.Contains(t, raw, "hide_ssid")
	assert.Contains(t, raw, "site_id")
}
