package exporter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clcollins/unifi-bootstrapper/internal/client"
	"github.com/clcollins/unifi-bootstrapper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testSite = "default"

// newMockExporter creates an Exporter with a fresh MockClient for testing.
func newMockExporter() (*Exporter, *client.MockClient) {
	mc := new(client.MockClient)
	e := NewExporter(mc, testSite)
	return e, mc
}

// setupAllSuccess configures all 8 mock endpoints to return the provided
// fixture data with no errors.
func setupAllSuccess(mc *client.MockClient, f fixtures) {
	mc.On("GetNetworks", mock.Anything, testSite).Return(f.networks, nil)
	mc.On("GetFirewallRules", mock.Anything, testSite).Return(f.firewallRules, nil)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return(f.firewallGroups, nil)
	mc.On("GetWLANs", mock.Anything, testSite).Return(f.wlans, nil)
	mc.On("GetPortForwards", mock.Anything, testSite).Return(f.portForwards, nil)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return(f.portProfiles, nil)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return(f.staticRoutes, nil)
	mc.On("GetDevices", mock.Anything, testSite).Return(f.devices, nil)
}

// fixtures holds sample data for all resource types used across tests.
type fixtures struct {
	networks       []models.Network
	firewallRules  []models.FirewallRule
	firewallGroups []models.FirewallGroup
	wlans          []models.WLAN
	portForwards   []models.PortForward
	portProfiles   []models.PortProfile
	staticRoutes   []models.StaticRoute
	devices        []models.Device
}

// sampleFixtures returns a set of non-empty test data for all resource types.
func sampleFixtures() fixtures {
	return fixtures{
		networks: []models.Network{
			{ID: "net-1", Name: "LAN", Purpose: "corporate", Subnet: "192.168.1.0/24"},
			{ID: "net-2", Name: "Guest", Purpose: "guest", Subnet: "192.168.2.0/24"},
		},
		firewallRules: []models.FirewallRule{
			{ID: "fw-1", Name: "Block IoT", Action: "drop", Enabled: true, Ruleset: "LAN_IN"},
		},
		firewallGroups: []models.FirewallGroup{
			{ID: "fwg-1", Name: "RFC1918", GroupType: "address-group", GroupMembers: []string{"10.0.0.0/8"}},
		},
		wlans: []models.WLAN{
			{ID: "wlan-1", Name: "HomeWiFi", Enabled: true, Security: "wpapsk"},
		},
		portForwards: []models.PortForward{
			{ID: "pf-1", Name: "SSH", Enabled: true, DstPort: "22", Fwd: "192.168.1.10", FwdPort: "22", Proto: "tcp"},
		},
		portProfiles: []models.PortProfile{
			{ID: "pp-1", Name: "All", Autoneg: true},
		},
		staticRoutes: []models.StaticRoute{
			{ID: "sr-1", Name: "VPN Route", Enabled: true, Type: "nexthop-route", StaticRouteNetwork: "10.0.0.0/8"},
		},
		devices: []models.Device{
			{ID: "dev-1", Name: "UDM-Pro", Model: "UDM", Type: "udm", Mac: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.1"},
		},
	}
}

// emptyFixtures returns fixture data with empty (non-nil) slices for all types.
func emptyFixtures() fixtures {
	return fixtures{
		networks:       []models.Network{},
		firewallRules:  []models.FirewallRule{},
		firewallGroups: []models.FirewallGroup{},
		wlans:          []models.WLAN{},
		portForwards:   []models.PortForward{},
		portProfiles:   []models.PortProfile{},
		staticRoutes:   []models.StaticRoute{},
		devices:        []models.Device{},
	}
}

func TestNewExporter(t *testing.T) {
	mc := new(client.MockClient)
	e := NewExporter(mc, "mysite")

	assert.NotNil(t, e)
	assert.Equal(t, "mysite", e.site)
}

func TestExport_AllSuccess(t *testing.T) {
	e, mc := newMockExporter()
	f := sampleFixtures()
	setupAllSuccess(mc, f)

	before := time.Now()
	inv, err := e.Export(context.Background())
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, inv)

	// Verify all resource types are populated.
	assert.Equal(t, f.networks, inv.Networks)
	assert.Equal(t, f.firewallRules, inv.FirewallRules)
	assert.Equal(t, f.firewallGroups, inv.FirewallGroups)
	assert.Equal(t, f.wlans, inv.WLANs)
	assert.Equal(t, f.portForwards, inv.PortForwards)
	assert.Equal(t, f.portProfiles, inv.PortProfiles)
	assert.Equal(t, f.staticRoutes, inv.StaticRoutes)
	assert.Equal(t, f.devices, inv.Devices)

	// Verify ExportedAt is set to approximately now.
	assert.False(t, inv.ExportedAt.IsZero(), "ExportedAt should be set")
	assert.True(t, inv.ExportedAt.After(before) || inv.ExportedAt.Equal(before),
		"ExportedAt should be at or after test start")
	assert.True(t, inv.ExportedAt.Before(after) || inv.ExportedAt.Equal(after),
		"ExportedAt should be at or before test end")

	mc.AssertExpectations(t)
}

func TestExport_PartialFailure_SingleEndpointErrors(t *testing.T) {
	e, mc := newMockExporter()
	f := sampleFixtures()

	// All succeed except GetNetworks, which returns an error.
	mc.On("GetNetworks", mock.Anything, testSite).Return([]models.Network(nil), errors.New("network timeout"))
	mc.On("GetFirewallRules", mock.Anything, testSite).Return(f.firewallRules, nil)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return(f.firewallGroups, nil)
	mc.On("GetWLANs", mock.Anything, testSite).Return(f.wlans, nil)
	mc.On("GetPortForwards", mock.Anything, testSite).Return(f.portForwards, nil)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return(f.portProfiles, nil)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return(f.staticRoutes, nil)
	mc.On("GetDevices", mock.Anything, testSite).Return(f.devices, nil)

	inv, err := e.Export(context.Background())

	// Should return both partial results AND the error.
	require.Error(t, err)
	require.NotNil(t, inv)
	assert.Contains(t, err.Error(), "network timeout")

	// Successful resources should still be present.
	assert.Equal(t, f.firewallRules, inv.FirewallRules)
	assert.Equal(t, f.firewallGroups, inv.FirewallGroups)
	assert.Equal(t, f.wlans, inv.WLANs)
	assert.Equal(t, f.portForwards, inv.PortForwards)
	assert.Equal(t, f.portProfiles, inv.PortProfiles)
	assert.Equal(t, f.staticRoutes, inv.StaticRoutes)
	assert.Equal(t, f.devices, inv.Devices)

	// Failed resource should be nil or empty.
	assert.Empty(t, inv.Networks)

	mc.AssertExpectations(t)
}

func TestExport_PartialFailure_MultipleEndpointsError(t *testing.T) {
	e, mc := newMockExporter()
	f := sampleFixtures()

	// Networks and Devices fail; others succeed.
	mc.On("GetNetworks", mock.Anything, testSite).Return([]models.Network(nil), errors.New("network error"))
	mc.On("GetFirewallRules", mock.Anything, testSite).Return(f.firewallRules, nil)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return(f.firewallGroups, nil)
	mc.On("GetWLANs", mock.Anything, testSite).Return(f.wlans, nil)
	mc.On("GetPortForwards", mock.Anything, testSite).Return(f.portForwards, nil)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return(f.portProfiles, nil)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return(f.staticRoutes, nil)
	mc.On("GetDevices", mock.Anything, testSite).Return([]models.Device(nil), errors.New("device error"))

	inv, err := e.Export(context.Background())

	require.Error(t, err)
	require.NotNil(t, inv)

	// Both error messages should be present.
	assert.Contains(t, err.Error(), "network error")
	assert.Contains(t, err.Error(), "device error")

	// Successful resources present.
	assert.Equal(t, f.firewallRules, inv.FirewallRules)
	assert.Equal(t, f.firewallGroups, inv.FirewallGroups)
	assert.Equal(t, f.wlans, inv.WLANs)
	assert.Equal(t, f.portForwards, inv.PortForwards)
	assert.Equal(t, f.portProfiles, inv.PortProfiles)
	assert.Equal(t, f.staticRoutes, inv.StaticRoutes)

	// Failed resources should be empty.
	assert.Empty(t, inv.Networks)
	assert.Empty(t, inv.Devices)

	mc.AssertExpectations(t)
}

func TestExport_AllFail(t *testing.T) {
	e, mc := newMockExporter()

	authErr := errors.New("authentication failed")
	mc.On("GetNetworks", mock.Anything, testSite).Return([]models.Network(nil), authErr)
	mc.On("GetFirewallRules", mock.Anything, testSite).Return([]models.FirewallRule(nil), authErr)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return([]models.FirewallGroup(nil), authErr)
	mc.On("GetWLANs", mock.Anything, testSite).Return([]models.WLAN(nil), authErr)
	mc.On("GetPortForwards", mock.Anything, testSite).Return([]models.PortForward(nil), authErr)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return([]models.PortProfile(nil), authErr)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return([]models.StaticRoute(nil), authErr)
	mc.On("GetDevices", mock.Anything, testSite).Return([]models.Device(nil), authErr)

	inv, err := e.Export(context.Background())

	require.Error(t, err)
	require.NotNil(t, inv)
	assert.Contains(t, err.Error(), "authentication failed")

	// All slices should be empty or nil.
	assert.Empty(t, inv.Networks)
	assert.Empty(t, inv.FirewallRules)
	assert.Empty(t, inv.FirewallGroups)
	assert.Empty(t, inv.WLANs)
	assert.Empty(t, inv.PortForwards)
	assert.Empty(t, inv.PortProfiles)
	assert.Empty(t, inv.StaticRoutes)
	assert.Empty(t, inv.Devices)

	mc.AssertExpectations(t)
}

func TestExport_EmptyResponses(t *testing.T) {
	e, mc := newMockExporter()
	f := emptyFixtures()
	setupAllSuccess(mc, f)

	inv, err := e.Export(context.Background())

	require.NoError(t, err)
	require.NotNil(t, inv)

	// All slices should be empty but not nil.
	assert.NotNil(t, inv.Networks)
	assert.Empty(t, inv.Networks)
	assert.NotNil(t, inv.FirewallRules)
	assert.Empty(t, inv.FirewallRules)
	assert.NotNil(t, inv.FirewallGroups)
	assert.Empty(t, inv.FirewallGroups)
	assert.NotNil(t, inv.WLANs)
	assert.Empty(t, inv.WLANs)
	assert.NotNil(t, inv.PortForwards)
	assert.Empty(t, inv.PortForwards)
	assert.NotNil(t, inv.PortProfiles)
	assert.Empty(t, inv.PortProfiles)
	assert.NotNil(t, inv.StaticRoutes)
	assert.Empty(t, inv.StaticRoutes)
	assert.NotNil(t, inv.Devices)
	assert.Empty(t, inv.Devices)

	// ExportedAt should still be set.
	assert.False(t, inv.ExportedAt.IsZero(), "ExportedAt should be set even for empty results")

	mc.AssertExpectations(t)
}

func TestExport_ContextCancellation(t *testing.T) {
	e, mc := newMockExporter()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	// Mock all endpoints to return context.Canceled when called with
	// an already-cancelled context.
	mc.On("GetNetworks", mock.Anything, testSite).Return([]models.Network(nil), context.Canceled)
	mc.On("GetFirewallRules", mock.Anything, testSite).Return([]models.FirewallRule(nil), context.Canceled)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return([]models.FirewallGroup(nil), context.Canceled)
	mc.On("GetWLANs", mock.Anything, testSite).Return([]models.WLAN(nil), context.Canceled)
	mc.On("GetPortForwards", mock.Anything, testSite).Return([]models.PortForward(nil), context.Canceled)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return([]models.PortProfile(nil), context.Canceled)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return([]models.StaticRoute(nil), context.Canceled)
	mc.On("GetDevices", mock.Anything, testSite).Return([]models.Device(nil), context.Canceled)

	inv, err := e.Export(ctx)

	require.Error(t, err)
	require.NotNil(t, inv)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestExport_ExportedAtIsSetAfterFetches(t *testing.T) {
	e, mc := newMockExporter()
	f := sampleFixtures()
	setupAllSuccess(mc, f)

	inv, err := e.Export(context.Background())

	require.NoError(t, err)
	require.NotNil(t, inv)

	// ExportedAt should be recent — within the last second.
	assert.WithinDuration(t, time.Now(), inv.ExportedAt, time.Second,
		"ExportedAt should be set to approximately now")
}

func TestExport_PartialFailure_ReturnsInventoryWithExportedAt(t *testing.T) {
	e, mc := newMockExporter()
	f := sampleFixtures()

	// Only WLANs fail.
	mc.On("GetNetworks", mock.Anything, testSite).Return(f.networks, nil)
	mc.On("GetFirewallRules", mock.Anything, testSite).Return(f.firewallRules, nil)
	mc.On("GetFirewallGroups", mock.Anything, testSite).Return(f.firewallGroups, nil)
	mc.On("GetWLANs", mock.Anything, testSite).Return([]models.WLAN(nil), errors.New("wlan fetch failed"))
	mc.On("GetPortForwards", mock.Anything, testSite).Return(f.portForwards, nil)
	mc.On("GetPortProfiles", mock.Anything, testSite).Return(f.portProfiles, nil)
	mc.On("GetStaticRoutes", mock.Anything, testSite).Return(f.staticRoutes, nil)
	mc.On("GetDevices", mock.Anything, testSite).Return(f.devices, nil)

	inv, err := e.Export(context.Background())

	require.Error(t, err)
	require.NotNil(t, inv)

	// ExportedAt should still be set on partial failure.
	assert.False(t, inv.ExportedAt.IsZero(), "ExportedAt should be set even on partial failure")

	mc.AssertExpectations(t)
}
