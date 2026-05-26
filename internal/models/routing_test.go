package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticRoute_UnmarshalFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/fixtures/static_routes.json")
	require.NoError(t, err)

	var resp APIResponse[StaticRoute]
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 2)

	// VPN route
	vpn := resp.Data[0]
	assert.Equal(t, "f2a3b4c5d6e7f8a9b0c1d2e3", vpn.ID)
	assert.Equal(t, "VPN Subnet Route", vpn.Name)
	assert.True(t, vpn.Enabled)
	assert.Equal(t, "nexthop-route", vpn.Type)
	assert.Equal(t, "172.16.10.0/24", vpn.StaticRouteNetwork)
	assert.Equal(t, "10.0.1.1", vpn.GatewayIP)
	assert.Equal(t, "ip-address", vpn.GatewayType)
	assert.Equal(t, 1, vpn.Distance)
	assert.Equal(t, "6a1b2c3d4e5f6a7b8c9d0e1f", vpn.SiteID)

	// Blackhole route
	blackhole := resp.Data[1]
	assert.Equal(t, "Blackhole Route", blackhole.Name)
	assert.Equal(t, "blackhole", blackhole.Type)
	assert.Equal(t, "192.168.200.0/24", blackhole.StaticRouteNetwork)
	assert.Equal(t, 254, blackhole.Distance)
}

func TestStaticRoute_RoundTrip(t *testing.T) {
	original := StaticRoute{
		ID:                 "abc123",
		Name:               "Test Route",
		Enabled:            true,
		Type:               "nexthop-route",
		StaticRouteNetwork: "172.16.20.0/24",
		GatewayIP:          "10.0.1.1",
		GatewayType:        "ip-address",
		Distance:           10,
		SiteID:             "site123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded StaticRoute
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestStaticRoute_OmitEmptyFields(t *testing.T) {
	route := StaticRoute{
		ID:                 "abc123",
		Name:               "Minimal",
		Enabled:            true,
		Type:               "blackhole",
		StaticRouteNetwork: "10.0.0.0/8",
		SiteID:             "site123",
	}

	data, err := json.Marshal(route)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "gateway_ip")
	assert.NotContains(t, raw, "gateway_type")
	assert.NotContains(t, raw, "interface")
	assert.NotContains(t, raw, "distance")
}

func TestStaticRoute_HyphenatedJSONKey(t *testing.T) {
	// The static-route_network key has a hyphen in it, which is
	// unusual. Verify it marshals and unmarshals correctly.
	route := StaticRoute{
		ID:                 "abc123",
		Name:               "Hyphen Test",
		Enabled:            true,
		Type:               "nexthop-route",
		StaticRouteNetwork: "10.99.0.0/16",
		SiteID:             "site123",
	}

	data, err := json.Marshal(route)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	assert.Equal(t, "10.99.0.0/16", raw["static-route_network"])
}
