package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortProfile_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/port_profiles.json")
	require.NoError(t, err)

	var resp APIResponse[PortProfile]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// All profile
	allProfile := resp.Data[0]
	assert.Equal(t, "d0e1f2a3b4c5d6e7f8a9b0c1", allProfile.ID)
	assert.Equal(t, "All", allProfile.Name)
	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", allProfile.NativeNetworkconfID)
	assert.Equal(t, []string{
		"6a3b4c5d6e7f8a9b0c1d2e3f",
		"7b4c5d6e7f8a9b0c1d2e3f4a",
	}, allProfile.TaggedNetworkconfIDs)
	assert.Equal(t, "auto", allProfile.PoeMode)
	assert.Equal(t, "all", allProfile.Forward)
	assert.True(t, allProfile.Autoneg)
	assert.True(t, allProfile.FullDuplex)
	assert.False(t, allProfile.StormctrlBcastEnabled)
	assert.False(t, allProfile.StormctrlMcastEnabled)
	assert.False(t, allProfile.StormctrlUcastEnabled)

	// Management Only profile
	mgmt := resp.Data[1]
	assert.Equal(t, "Management Only", mgmt.Name)
	assert.Equal(t, "off", mgmt.PoeMode)
	assert.Equal(t, "native", mgmt.Forward)
	assert.Equal(t, 1000, mgmt.Speed)
	assert.True(t, mgmt.StormctrlBcastEnabled)
	assert.True(t, mgmt.StormctrlMcastEnabled)
	assert.False(t, mgmt.StormctrlUcastEnabled)
	assert.Empty(t, mgmt.TaggedNetworkconfIDs)
}

func TestPortProfile_RoundTrip(t *testing.T) {
	original := PortProfile{
		ID:                    "abc123",
		Name:                  "Test Profile",
		NativeNetworkconfID:   "net123",
		TaggedNetworkconfIDs:  []string{"net456", "net789"},
		PoeMode:               "auto",
		Forward:               "all",
		SiteID:                "site123",
		Autoneg:               true,
		Speed:                 1000,
		FullDuplex:            true,
		StormctrlBcastEnabled: true,
		StormctrlMcastEnabled: false,
		StormctrlUcastEnabled: false,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded PortProfile
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestPortProfile_OmitEmptyFields(t *testing.T) {
	pp := PortProfile{
		ID:                   "abc123",
		Name:                 "Minimal",
		TaggedNetworkconfIDs: []string{},
		SiteID:               "site123",
	}

	data, err := json.Marshal(pp)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "native_networkconf_id")
	assert.NotContains(t, raw, "poe_mode")
	assert.NotContains(t, raw, "forward")
	assert.NotContains(t, raw, "speed")
}

func TestPortProfile_EmptyTaggedNetworks(t *testing.T) {
	// Verify that an empty TaggedNetworkconfIDs slice marshals
	// as [] rather than null
	pp := PortProfile{
		ID:                   "abc123",
		Name:                 "Empty Tags",
		TaggedNetworkconfIDs: []string{},
		SiteID:               "site123",
	}

	data, err := json.Marshal(pp)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	taggedRaw, ok := raw["tagged_networkconf_ids"]
	require.True(t, ok)
	tagged, ok := taggedRaw.([]interface{})
	require.True(t, ok)
	assert.Empty(t, tagged)
}
