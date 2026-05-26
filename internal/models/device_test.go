package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevice_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/devices.json")
	require.NoError(t, err)

	var resp APIResponse[Device]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// UDM-Pro
	udm := resp.Data[0]
	assert.Equal(t, "b4c5d6e7f8a9b0c1d2e3f4a5", udm.ID)
	assert.Equal(t, "UDM-Pro", udm.Name)
	assert.Equal(t, "UDM", udm.Model)
	assert.Equal(t, "UniFi Dream Machine Pro", udm.ModelName)
	assert.Equal(t, "udm", udm.Type)
	assert.Equal(t, "aa:bb:cc:dd:ee:01", udm.Mac)
	assert.Equal(t, "10.0.1.1", udm.IP)
	assert.Equal(t, "3.2.7.2", udm.Version)
	assert.Equal(t, "FAKE0000000001", udm.Serial)
	assert.True(t, udm.Adopted)
	assert.Equal(t, 1, udm.State)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", udm.SiteID)
	assert.Equal(t, 864000, udm.Uptime)

	// Switch
	sw := resp.Data[1]
	assert.Equal(t, "Office Switch 24", sw.Name)
	assert.Equal(t, "US24", sw.Model)
	assert.Equal(t, "usw", sw.Type)
	assert.Equal(t, "aa:bb:cc:dd:ee:02", sw.Mac)
	assert.Equal(t, 1, sw.State)
}

func TestDevice_RoundTrip(t *testing.T) {
	original := Device{
		ID:        "abc123",
		Name:      "Test Device",
		Model:     "UAP",
		ModelName: "UniFi AP",
		Type:      "uap",
		Mac:       "ff:ee:dd:cc:bb:aa",
		IP:        "10.0.1.50",
		Version:   "6.0.0",
		Serial:    "TESTSERIAL",
		Adopted:   true,
		State:     1,
		SiteID:    "site123",
		Uptime:    3600,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Device
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDevice_OmitEmptyFields(t *testing.T) {
	dev := Device{
		ID:      "abc123",
		Name:    "Minimal",
		Model:   "UDM",
		Type:    "udm",
		Mac:     "aa:bb:cc:dd:ee:ff",
		IP:      "10.0.1.1",
		Adopted: true,
		State:   1,
		SiteID:  "site123",
	}

	data, err := json.Marshal(dev)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "model_name")
	assert.NotContains(t, raw, "version")
	assert.NotContains(t, raw, "serial")
	assert.NotContains(t, raw, "uptime")

	// Non-omitempty fields present
	assert.Contains(t, raw, "_id")
	assert.Contains(t, raw, "name")
	assert.Contains(t, raw, "model")
	assert.Contains(t, raw, "type")
	assert.Contains(t, raw, "mac")
	assert.Contains(t, raw, "ip")
	assert.Contains(t, raw, "adopted")
	assert.Contains(t, raw, "state")
	assert.Contains(t, raw, "site_id")
}
