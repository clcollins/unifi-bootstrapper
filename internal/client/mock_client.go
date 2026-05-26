package client

import (
	"context"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockClient is a testify/mock implementation of ClientInterface for
// use in tests of downstream packages (exporter, generator, renderer).
type MockClient struct {
	mock.Mock
}

// Ping mocks the Ping method.
func (m *MockClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// GetNetworks mocks the GetNetworks method.
func (m *MockClient) GetNetworks(ctx context.Context, site string) ([]models.Network, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.Network), args.Error(1)
}

// GetFirewallRules mocks the GetFirewallRules method.
func (m *MockClient) GetFirewallRules(ctx context.Context, site string) ([]models.FirewallRule, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.FirewallRule), args.Error(1)
}

// GetFirewallGroups mocks the GetFirewallGroups method.
func (m *MockClient) GetFirewallGroups(ctx context.Context, site string) ([]models.FirewallGroup, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.FirewallGroup), args.Error(1)
}

// GetWLANs mocks the GetWLANs method.
func (m *MockClient) GetWLANs(ctx context.Context, site string) ([]models.WLAN, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.WLAN), args.Error(1)
}

// GetPortForwards mocks the GetPortForwards method.
func (m *MockClient) GetPortForwards(ctx context.Context, site string) ([]models.PortForward, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.PortForward), args.Error(1)
}

// GetPortProfiles mocks the GetPortProfiles method.
func (m *MockClient) GetPortProfiles(ctx context.Context, site string) ([]models.PortProfile, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.PortProfile), args.Error(1)
}

// GetStaticRoutes mocks the GetStaticRoutes method.
func (m *MockClient) GetStaticRoutes(ctx context.Context, site string) ([]models.StaticRoute, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.StaticRoute), args.Error(1)
}

// GetDevices mocks the GetDevices method.
func (m *MockClient) GetDevices(ctx context.Context, site string) ([]models.Device, error) {
	args := m.Called(ctx, site)
	return args.Get(0).([]models.Device), args.Error(1)
}
