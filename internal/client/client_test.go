package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// projectRoot returns the absolute path to the project root by walking up
// from the test file's directory until it finds go.mod.
func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get caller info")

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "go.mod not found in any parent directory")
		dir = parent
	}
}

// loadFixture reads a fixture file from testdata/fixtures/ and returns
// its contents as bytes.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(projectRoot(t), "testdata", "fixtures", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read fixture: %s", name)
	return data
}

// newTLSTestServer creates an httptest TLS server with the given handler
// and returns the server and a Client configured to connect to it.
func newTLSTestServer(t *testing.T, handler http.Handler, opts ...Option) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	allOpts := []Option{
		WithHTTPClient(server.Client()),
		WithInsecure(true),
	}
	allOpts = append(allOpts, opts...)

	c := NewClient(server.URL, allOpts...)
	return server, c
}

// --- Ping tests ---

func TestPing_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/stat/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	err := c.Ping(context.Background())
	assert.NoError(t, err)
}

func TestPing_AuthFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"meta":{"rc":"error","msg":"api.err.Unauthorized"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	err := c.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestPing_ConnectionFailure(t *testing.T) {
	// Create a server and immediately close it so the port is refused.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := server.URL
	server.Close()

	c := NewClient(url, WithInsecure(true))
	err := c.Ping(context.Background())
	require.Error(t, err)
}

// --- API key auth tests ---

func TestAPIKeyAuth_HeaderSet(t *testing.T) {
	const fakeAPIKey = "test-api-key-not-real"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fakeAPIKey, r.Header.Get("X-API-KEY"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler, WithAPIKey(fakeAPIKey))
	err := c.Ping(context.Background())
	assert.NoError(t, err)
}

// --- Cookie-session auth tests ---

func TestCookieSessionAuth_LoginAndCookies(t *testing.T) {
	const (
		fakeUser = "testadmin"
		fakePass = "testpassword"
		fakeCSRF = "csrf-token-fake-value"
	)

	loginCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy/network/api/login" && r.Method == http.MethodPost {
			loginCalled = true

			var creds struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			err := json.NewDecoder(r.Body).Decode(&creds)
			require.NoError(t, err)
			assert.Equal(t, fakeUser, creds.Username)
			assert.Equal(t, fakePass, creds.Password)

			http.SetCookie(w, &http.Cookie{
				Name:  "TOKEN",
				Value: fakeCSRF,
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "unifises",
				Value: "session-id-fake",
			})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
			return
		}

		// Subsequent requests: verify cookies are present and CSRF header set
		var hasSessionCookie bool
		for _, cookie := range r.Cookies() {
			if cookie.Name == "unifises" {
				hasSessionCookie = true
			}
		}
		assert.True(t, hasSessionCookie, "session cookie should be present on subsequent requests")
		assert.Equal(t, fakeCSRF, r.Header.Get("X-CSRF-Token"), "CSRF token header should be set")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler, WithCredentials(fakeUser, fakePass))
	err := c.Ping(context.Background())
	assert.NoError(t, err)
	assert.True(t, loginCalled, "login endpoint should have been called")
}

// --- Insecure TLS test ---

func TestInsecureTLS_ConnectsToSelfSigned(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	// Use a fresh HTTP client (not server.Client()) to verify the
	// insecure option actually skips TLS verification.
	c := NewClient(server.URL, WithInsecure(true))
	err := c.Ping(context.Background())
	assert.NoError(t, err)
}

func TestSecureTLS_RejectsSelfSigned(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	// Default client without insecure or injected TLS config should
	// reject the self-signed certificate.
	c := NewClient(server.URL)
	err := c.Ping(context.Background())
	require.Error(t, err)
}

// --- HTTP error handling ---

func TestHTTPErrorHandling_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	})

	_, c := newTLSTestServer(t, handler)
	err := c.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// --- API error handling ---

func TestAPIErrorHandling_ErrorResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"error","msg":"api.err.Unauthorized"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	_, err := c.GetNetworks(context.Background(), "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api.err.Unauthorized")
}

// --- GetNetworks ---

func TestGetNetworks_Success(t *testing.T) {
	fixture := loadFixture(t, "networks.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/networkconf", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	networks, err := c.GetNetworks(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, networks, 3)

	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", networks[0].ID)
	assert.Equal(t, "Default LAN", networks[0].Name)
	assert.Equal(t, "corporate", networks[0].Purpose)
	assert.Equal(t, "10.0.1.0/24", networks[0].Subnet)
	assert.False(t, networks[0].VLANEnabled)
	assert.True(t, networks[0].DHCPEnabled)
	assert.Equal(t, "10.0.1.100", networks[0].DHCPStart)
	assert.Equal(t, "10.0.1.254", networks[0].DHCPStop)
	assert.True(t, networks[0].DHCPDNSEnabled)
	assert.Equal(t, "localdomain", networks[0].DomainName)
	assert.True(t, networks[0].IGMPSnooping)
	assert.Equal(t, "LAN", networks[0].NetworkGroup)

	// Guest network with VLAN
	assert.Equal(t, "Guest Network", networks[1].Name)
	assert.Equal(t, "guest", networks[1].Purpose)
	assert.Equal(t, 50, networks[1].VLANID)
	assert.True(t, networks[1].VLANEnabled)

	// Management VLAN
	assert.Equal(t, "Management VLAN", networks[2].Name)
	assert.Equal(t, 99, networks[2].VLANID)
	assert.True(t, networks[2].VLANEnabled)
}

// --- GetFirewallRules ---

func TestGetFirewallRules_Success(t *testing.T) {
	fixture := loadFixture(t, "firewall_rules.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/firewallrule", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	rules, err := c.GetFirewallRules(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, rules, 3)

	// First rule has rule_index as string "2000"
	assert.Equal(t, "Allow Established", rules[0].Name)
	assert.Equal(t, "accept", rules[0].Action)
	assert.Equal(t, 2000, rules[0].RuleIndex.Int())
	assert.Equal(t, "LAN_IN", rules[0].Ruleset)
	assert.True(t, rules[0].StateEstablished)
	assert.True(t, rules[0].StateRelated)

	// Second rule has rule_index as int 2001
	assert.Equal(t, "Drop Invalid", rules[1].Name)
	assert.Equal(t, 2001, rules[1].RuleIndex.Int())
	assert.True(t, rules[1].StateInvalid)
	assert.True(t, rules[1].Logging)

	// Third rule
	assert.Equal(t, "Block Guest to LAN", rules[2].Name)
	assert.Equal(t, "reject", rules[2].Action)
	assert.Equal(t, 3000, rules[2].RuleIndex.Int())
	assert.Equal(t, "GUEST_IN", rules[2].Ruleset)
}

// --- GetFirewallGroups ---

func TestGetFirewallGroups_Success(t *testing.T) {
	fixture := loadFixture(t, "firewall_groups.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/firewallgroup", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	groups, err := c.GetFirewallGroups(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, groups, 2)

	assert.Equal(t, "RFC1918 Addresses", groups[0].Name)
	assert.Equal(t, "address-group", groups[0].GroupType)
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}, groups[0].GroupMembers)

	assert.Equal(t, "Web Ports", groups[1].Name)
	assert.Equal(t, "port-group", groups[1].GroupType)
	assert.Equal(t, []string{"80", "443", "8080", "8443"}, groups[1].GroupMembers)
}

// --- GetWLANs ---

func TestGetWLANs_Success(t *testing.T) {
	fixture := loadFixture(t, "wlans.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/wlanconf", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	wlans, err := c.GetWLANs(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, wlans, 2)

	assert.Equal(t, "HomeNet-5G", wlans[0].Name)
	assert.True(t, wlans[0].Enabled)
	assert.Equal(t, "wpapsk", wlans[0].Security)
	assert.Equal(t, "wpa2", wlans[0].WPAMode)
	assert.Equal(t, "test-passphrase-do-not-use", wlans[0].XPassphrase)
	assert.False(t, wlans[0].VLANEnabled)
	assert.False(t, wlans[0].IsGuest)
	assert.Equal(t, "both", wlans[0].WlanBand)

	assert.Equal(t, "GuestWiFi", wlans[1].Name)
	assert.True(t, wlans[1].VLANEnabled)
	assert.Equal(t, 50, wlans[1].VLANID)
	assert.True(t, wlans[1].IsGuest)
	assert.Equal(t, "guest-fake-password-not-real", wlans[1].XPassphrase)
}

// --- GetPortForwards ---

func TestGetPortForwards_Success(t *testing.T) {
	fixture := loadFixture(t, "port_forwards.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/portforward", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	forwards, err := c.GetPortForwards(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, forwards, 2)

	assert.Equal(t, "Web Server HTTP", forwards[0].Name)
	assert.True(t, forwards[0].Enabled)
	assert.Equal(t, "80", forwards[0].DstPort)
	assert.Equal(t, "10.0.1.200", forwards[0].Fwd)
	assert.Equal(t, "8080", forwards[0].FwdPort)
	assert.Equal(t, "tcp", forwards[0].Proto)
	assert.False(t, forwards[0].Log)

	assert.Equal(t, "Minecraft Server", forwards[1].Name)
	assert.Equal(t, "tcp_udp", forwards[1].Proto)
	assert.True(t, forwards[1].Log)
}

// --- GetPortProfiles ---

func TestGetPortProfiles_Success(t *testing.T) {
	fixture := loadFixture(t, "port_profiles.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/portconf", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	profiles, err := c.GetPortProfiles(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, profiles, 2)

	assert.Equal(t, "All", profiles[0].Name)
	assert.Equal(t, "5f2e4a3b1c9d8e7f6a5b4c3d", profiles[0].NativeNetworkconfID)
	assert.Equal(t, []string{"6a3b4c5d6e7f8a9b0c1d2e3f", "7b4c5d6e7f8a9b0c1d2e3f4a"}, profiles[0].TaggedNetworkconfIDs)
	assert.Equal(t, "auto", profiles[0].PoeMode)
	assert.True(t, profiles[0].Autoneg)
	assert.True(t, profiles[0].FullDuplex)

	assert.Equal(t, "Management Only", profiles[1].Name)
	assert.Equal(t, "off", profiles[1].PoeMode)
	assert.Equal(t, 1000, profiles[1].Speed)
	assert.True(t, profiles[1].StormctrlBcastEnabled)
	assert.True(t, profiles[1].StormctrlMcastEnabled)
	assert.False(t, profiles[1].StormctrlUcastEnabled)
}

// --- GetStaticRoutes ---

func TestGetStaticRoutes_Success(t *testing.T) {
	fixture := loadFixture(t, "static_routes.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/rest/routing", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	routes, err := c.GetStaticRoutes(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, routes, 2)

	assert.Equal(t, "VPN Subnet Route", routes[0].Name)
	assert.True(t, routes[0].Enabled)
	assert.Equal(t, "nexthop-route", routes[0].Type)
	assert.Equal(t, "172.16.10.0/24", routes[0].StaticRouteNetwork)
	assert.Equal(t, "10.0.1.1", routes[0].GatewayIP)
	assert.Equal(t, 1, routes[0].Distance)

	assert.Equal(t, "Blackhole Route", routes[1].Name)
	assert.Equal(t, "blackhole", routes[1].Type)
	assert.Equal(t, 254, routes[1].Distance)
}

// --- GetDevices ---

func TestGetDevices_Success(t *testing.T) {
	fixture := loadFixture(t, "devices.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/default/stat/device", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	})

	_, c := newTLSTestServer(t, handler)
	devices, err := c.GetDevices(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, devices, 2)

	assert.Equal(t, "UDM-Pro", devices[0].Name)
	assert.Equal(t, "UDM", devices[0].Model)
	assert.Equal(t, "UniFi Dream Machine Pro", devices[0].ModelName)
	assert.Equal(t, "udm", devices[0].Type)
	assert.Equal(t, "aa:bb:cc:dd:ee:01", devices[0].Mac)
	assert.Equal(t, "10.0.1.1", devices[0].IP)
	assert.Equal(t, "3.2.7.2", devices[0].Version)
	assert.Equal(t, "FAKE0000000001", devices[0].Serial)
	assert.True(t, devices[0].Adopted)
	assert.Equal(t, 1, devices[0].State)
	assert.Equal(t, 864000, devices[0].Uptime)

	assert.Equal(t, "Office Switch 24", devices[1].Name)
	assert.Equal(t, "US24", devices[1].Model)
}

// --- Site parameter in paths ---

func TestSiteParameter_CustomSite(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proxy/network/api/s/mysiteabc/rest/networkconf", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	networks, err := c.GetNetworks(context.Background(), "mysiteabc")
	require.NoError(t, err)
	assert.Empty(t, networks)
}

// --- Path prefix ---

func TestAllPathsPrefixed_ProxyNetwork(t *testing.T) {
	var requestedPaths []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	ctx := context.Background()

	_ = c.Ping(ctx)
	_, _ = c.GetNetworks(ctx, "default")
	_, _ = c.GetFirewallRules(ctx, "default")
	_, _ = c.GetFirewallGroups(ctx, "default")
	_, _ = c.GetWLANs(ctx, "default")
	_, _ = c.GetPortForwards(ctx, "default")
	_, _ = c.GetPortProfiles(ctx, "default")
	_, _ = c.GetStaticRoutes(ctx, "default")
	_, _ = c.GetDevices(ctx, "default")

	for _, path := range requestedPaths {
		assert.True(t, strings.HasPrefix(path, "/proxy/network"),
			"path %q should start with /proxy/network", path)
	}
	assert.Len(t, requestedPaths, 9, "expected 9 API calls")
}

// --- Empty data array ---

func TestGetNetworks_EmptyData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	networks, err := c.GetNetworks(context.Background(), "default")
	require.NoError(t, err)
	assert.Empty(t, networks)
}

// --- ClientInterface compliance ---

func TestClient_ImplementsClientInterface(t *testing.T) {
	var _ ClientInterface = (*Client)(nil)
}

// --- Cookie-session auth login failure ---

func TestCookieSessionAuth_LoginFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy/network/api/login" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"meta":{"rc":"error","msg":"api.err.InvalidCredentials"},"data":[]}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler, WithCredentials("bad", "creds"))
	err := c.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login")
}

// --- Verify all Get methods return correct types ---

func TestGetMethods_ReturnTypes(t *testing.T) {
	fixtures := map[string]string{
		"/proxy/network/api/s/default/rest/networkconf":  "networks.json",
		"/proxy/network/api/s/default/rest/firewallrule": "firewall_rules.json",
		"/proxy/network/api/s/default/rest/firewallgroup": "firewall_groups.json",
		"/proxy/network/api/s/default/rest/wlanconf":     "wlans.json",
		"/proxy/network/api/s/default/rest/portforward":  "port_forwards.json",
		"/proxy/network/api/s/default/rest/portconf":     "port_profiles.json",
		"/proxy/network/api/s/default/rest/routing":      "static_routes.json",
		"/proxy/network/api/s/default/stat/device":       "devices.json",
		"/proxy/network/api/s/default/stat/health":       "",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixtureName, ok := fixtures[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if fixtureName == "" {
			_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
			return
		}
		_, _ = w.Write(loadFixture(t, fixtureName))
	})

	_, c := newTLSTestServer(t, handler)
	ctx := context.Background()

	networks, err := c.GetNetworks(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.Network{}, networks)

	rules, err := c.GetFirewallRules(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.FirewallRule{}, rules)

	groups, err := c.GetFirewallGroups(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.FirewallGroup{}, groups)

	wlans, err := c.GetWLANs(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.WLAN{}, wlans)

	forwards, err := c.GetPortForwards(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.PortForward{}, forwards)

	profiles, err := c.GetPortProfiles(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.PortProfile{}, profiles)

	routes, err := c.GetStaticRoutes(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.StaticRoute{}, routes)

	devices, err := c.GetDevices(ctx, "default")
	require.NoError(t, err)
	assert.IsType(t, []models.Device{}, devices)
}

// --- Context cancellation ---

func TestContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := c.Ping(ctx)
	require.Error(t, err)
}

// --- NewClient defaults ---

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("https://192.168.1.1")
	assert.NotNil(t, c)
	assert.Equal(t, "https://192.168.1.1", c.host)
}

// --- Multiple options ---

func TestNewClient_MultipleOptions(t *testing.T) {
	c := NewClient("https://192.168.1.1",
		WithAPIKey("test-key"),
		WithInsecure(true),
	)
	assert.NotNil(t, c)
	assert.Equal(t, "test-key", c.apiKey)
	assert.True(t, c.insecure)
}

// --- Invalid JSON response ---

func TestGetNetworks_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	})

	_, c := newTLSTestServer(t, handler)
	_, err := c.GetNetworks(context.Background(), "default")
	require.Error(t, err)
}

// --- Mock client tests ---

func TestMockClient_ImplementsInterface(t *testing.T) {
	var _ ClientInterface = (*MockClient)(nil)
}

func TestMockClient_GetNetworks(t *testing.T) {
	m := new(MockClient)
	expected := []models.Network{
		{ID: "abc123", Name: "Test Network"},
	}
	m.On("GetNetworks", context.Background(), "default").Return(expected, nil)

	result, err := m.GetNetworks(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetNetworks_Error(t *testing.T) {
	m := new(MockClient)
	m.On("GetNetworks", context.Background(), "default").Return([]models.Network(nil), fmt.Errorf("connection refused"))

	result, err := m.GetNetworks(context.Background(), "default")
	require.Error(t, err)
	assert.Nil(t, result)
	m.AssertExpectations(t)
}

func TestMockClient_Ping(t *testing.T) {
	m := new(MockClient)
	m.On("Ping", context.Background()).Return(nil)

	err := m.Ping(context.Background())
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestMockClient_GetFirewallRules(t *testing.T) {
	m := new(MockClient)
	expected := []models.FirewallRule{{ID: "r1", Name: "Test Rule"}}
	m.On("GetFirewallRules", context.Background(), "default").Return(expected, nil)

	result, err := m.GetFirewallRules(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetFirewallGroups(t *testing.T) {
	m := new(MockClient)
	expected := []models.FirewallGroup{{ID: "g1", Name: "Test Group"}}
	m.On("GetFirewallGroups", context.Background(), "default").Return(expected, nil)

	result, err := m.GetFirewallGroups(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetWLANs(t *testing.T) {
	m := new(MockClient)
	expected := []models.WLAN{{ID: "w1", Name: "Test WLAN"}}
	m.On("GetWLANs", context.Background(), "default").Return(expected, nil)

	result, err := m.GetWLANs(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetPortForwards(t *testing.T) {
	m := new(MockClient)
	expected := []models.PortForward{{ID: "pf1", Name: "Test Forward"}}
	m.On("GetPortForwards", context.Background(), "default").Return(expected, nil)

	result, err := m.GetPortForwards(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetPortProfiles(t *testing.T) {
	m := new(MockClient)
	expected := []models.PortProfile{{ID: "pp1", Name: "Test Profile"}}
	m.On("GetPortProfiles", context.Background(), "default").Return(expected, nil)

	result, err := m.GetPortProfiles(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetStaticRoutes(t *testing.T) {
	m := new(MockClient)
	expected := []models.StaticRoute{{ID: "sr1", Name: "Test Route"}}
	m.On("GetStaticRoutes", context.Background(), "default").Return(expected, nil)

	result, err := m.GetStaticRoutes(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockClient_GetDevices(t *testing.T) {
	m := new(MockClient)
	expected := []models.Device{{ID: "d1", Name: "Test Device"}}
	m.On("GetDevices", context.Background(), "default").Return(expected, nil)

	result, err := m.GetDevices(context.Background(), "default")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

// --- truncateBody coverage ---

func TestTruncateBody_LongBody(t *testing.T) {
	// Build a body longer than 200 bytes to exercise the truncation path.
	long := strings.Repeat("x", 300)
	result := truncateBody([]byte(long))
	assert.Len(t, result, 203) // 200 + "..."
	assert.True(t, strings.HasSuffix(result, "..."))
}

// --- doGet read body error is difficult to test, but we can cover
// the login marshal error path indirectly and the request creation
// error path ---

func TestLogin_Success_CookiesExtracted(t *testing.T) {
	// Verify that login sets the csrfToken and loggedIn fields.
	const fakeCSRF = "extracted-csrf-token"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy/network/api/login" {
			http.SetCookie(w, &http.Cookie{
				Name:  "TOKEN",
				Value: fakeCSRF,
			})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
			return
		}
		// Verify CSRF token is used on subsequent request.
		assert.Equal(t, fakeCSRF, r.Header.Get("X-CSRF-Token"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler, WithCredentials("user", "pass"))
	_, err := c.GetNetworks(context.Background(), "default")
	require.NoError(t, err)
	assert.True(t, c.loggedIn)
	assert.Equal(t, fakeCSRF, c.csrfToken)
}

// --- API error with no msg field ---

func TestAPIError_NoMsgField(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"rc":"error"},"data":[]}`))
	})

	_, c := newTLSTestServer(t, handler)
	_, err := c.GetNetworks(context.Background(), "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error")
}
