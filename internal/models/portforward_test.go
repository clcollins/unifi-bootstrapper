package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortForward_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/port_forwards.json")
	require.NoError(t, err)

	var resp APIResponse[PortForward]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// Web server forward
	web := resp.Data[0]
	assert.Equal(t, "b8c9d0e1f2a3b4c5d6e7f8a9", web.ID)
	assert.Equal(t, "Web Server HTTP", web.Name)
	assert.True(t, web.Enabled)
	assert.Equal(t, "any", web.Src)
	assert.Equal(t, "80", web.DstPort)
	assert.Equal(t, "10.0.1.200", web.Fwd)
	assert.Equal(t, "8080", web.FwdPort)
	assert.Equal(t, "tcp", web.Proto)
	assert.False(t, web.Log)
	assert.Equal(t, "wan", web.PfwdInterface)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", web.SiteID)

	// Minecraft forward
	mc := resp.Data[1]
	assert.Equal(t, "Minecraft Server", mc.Name)
	assert.Equal(t, "25565", mc.DstPort)
	assert.Equal(t, "10.0.1.150", mc.Fwd)
	assert.Equal(t, "tcp_udp", mc.Proto)
	assert.True(t, mc.Log)
}

func TestPortForward_RoundTrip(t *testing.T) {
	original := PortForward{
		ID:            "abc123",
		Name:          "Test Forward",
		Enabled:       true,
		Src:           "any",
		DstPort:       "8080",
		Fwd:           "10.0.1.100",
		FwdPort:       "80",
		Proto:         "tcp",
		Log:           true,
		SiteID:        "site123",
		PfwdInterface: "wan",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded PortForward
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestPortForward_OmitEmptyFields(t *testing.T) {
	pf := PortForward{
		ID:      "abc123",
		Name:    "Minimal",
		Enabled: true,
		DstPort: "80",
		Fwd:     "10.0.1.1",
		FwdPort: "80",
		Proto:   "tcp",
		SiteID:  "site123",
	}

	data, err := json.Marshal(pf)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "src")
	assert.NotContains(t, raw, "pfwd_interface")
}
