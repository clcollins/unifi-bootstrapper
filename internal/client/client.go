// Package client provides an HTTP client for communicating with the
// UniFi UDM-Pro local API. It supports API key authentication and
// cookie-session authentication, and handles TLS verification settings.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

const (
	// pathPrefix is prepended to all API paths for UDM-Pro compatibility.
	pathPrefix = "/proxy/network"

	// loginPath is the cookie-session authentication endpoint.
	loginPath = pathPrefix + "/api/login"
)

// apiErrorResponse is used internally to parse the meta.msg field from
// API error responses. The models.APIResponse type only exposes meta.rc,
// so this struct captures the additional msg field for error messages.
type apiErrorResponse struct {
	Meta struct {
		RC  string `json:"rc"`
		Msg string `json:"msg"`
	} `json:"meta"`
}

// ClientInterface defines the methods for fetching UniFi resources.
// Both the real Client and MockClient implement this interface.
type ClientInterface interface {
	Ping(ctx context.Context) error
	GetNetworks(ctx context.Context, site string) ([]models.Network, error)
	GetFirewallRules(ctx context.Context, site string) ([]models.FirewallRule, error)
	GetFirewallGroups(ctx context.Context, site string) ([]models.FirewallGroup, error)
	GetWLANs(ctx context.Context, site string) ([]models.WLAN, error)
	GetPortForwards(ctx context.Context, site string) ([]models.PortForward, error)
	GetPortProfiles(ctx context.Context, site string) ([]models.PortProfile, error)
	GetStaticRoutes(ctx context.Context, site string) ([]models.StaticRoute, error)
	GetDevices(ctx context.Context, site string) ([]models.Device, error)
}

// Client is the real HTTP client for communicating with a UDM-Pro API.
type Client struct {
	host       string
	apiKey     string
	username   string
	password   string
	insecure   bool
	httpClient *http.Client
	csrfToken  string
	loggedIn   bool
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithAPIKey sets API key authentication. The key is sent as an
// X-API-KEY header on every request.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithCredentials sets cookie-session authentication. The client will
// POST to the login endpoint to obtain session cookies before making
// API requests.
func WithCredentials(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithInsecure controls TLS certificate verification. When true, the
// client skips TLS verification, which is necessary for connecting to
// UDM-Pro devices with self-signed certificates.
func WithInsecure(insecure bool) Option {
	return func(c *Client) {
		c.insecure = insecure
	}
}

// WithHTTPClient injects a custom *http.Client, useful for testing
// with httptest servers.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Client for the given host with the provided
// options applied.
func NewClient(host string, opts ...Option) *Client {
	c := &Client{
		host: host,
	}

	for _, opt := range opts {
		opt(c)
	}

	// If no custom HTTP client was injected, create one with the
	// appropriate TLS configuration.
	if c.httpClient == nil {
		transport := &http.Transport{}
		if c.insecure {
			transport.TLSClientConfig = &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec // user-requested insecure mode
			}
		}
		c.httpClient = &http.Client{
			Transport: transport,
		}
	}

	// Set up a cookie jar for session cookie management.
	jar, _ := cookiejar.New(nil)
	c.httpClient.Jar = jar

	return c
}

// login performs cookie-session authentication by POSTing credentials
// to the login endpoint. It captures session cookies and extracts the
// CSRF token from the TOKEN cookie for use in subsequent requests.
func (c *Client) login(ctx context.Context) error {
	creds, err := json.Marshal(map[string]string{
		"username": c.username,
		"password": c.password,
	})
	if err != nil {
		return fmt.Errorf("marshaling login credentials: %w", err)
	}

	url := c.host + loginPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(creds))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with HTTP status %d", resp.StatusCode)
	}

	// Extract CSRF token from the TOKEN cookie.
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "TOKEN" {
			c.csrfToken = cookie.Value
		}
	}

	c.loggedIn = true
	return nil
}

// doGet performs an authenticated GET request to the given path, checks
// for HTTP and API-level errors, and returns the raw response body.
func (c *Client) doGet(ctx context.Context, path string) ([]byte, error) {
	// If using cookie-session auth and not yet logged in, log in first.
	if c.username != "" && !c.loggedIn {
		if err := c.login(ctx); err != nil {
			return nil, fmt.Errorf("login required but failed: %w", err)
		}
	}

	url := c.host + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", path, err)
	}

	// Set authentication headers.
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}
	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body from %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, path, truncateBody(body))
	}

	return body, nil
}

// fetch performs a GET request to the given path, unmarshals the
// APIResponse envelope, checks for API-level errors, and returns the
// data slice.
func fetch[T any](c *Client, ctx context.Context, path string) ([]T, error) {
	body, err := c.doGet(ctx, path)
	if err != nil {
		return nil, err
	}

	// Check for API-level error first (meta.rc != "ok").
	var errResp apiErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Meta.RC != "ok" {
			msg := errResp.Meta.Msg
			if msg == "" {
				msg = errResp.Meta.RC
			}
			return nil, fmt.Errorf("API error from %s: %s", path, msg)
		}
	}

	var apiResp models.APIResponse[T]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response from %s: %w", path, err)
	}

	return apiResp.Data, nil
}

// truncateBody returns the first 200 bytes of a response body for
// inclusion in error messages.
func truncateBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// sitePath builds an API path for the given site and resource suffix.
func sitePath(site, suffix string) string {
	return fmt.Sprintf("%s/api/s/%s/%s", pathPrefix, site, suffix)
}

// Ping checks connectivity and authentication by hitting the site
// health endpoint.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.doGet(ctx, sitePath("default", "stat/health"))
	return err
}

// GetNetworks fetches all network configurations for the given site.
func (c *Client) GetNetworks(ctx context.Context, site string) ([]models.Network, error) {
	return fetch[models.Network](c, ctx, sitePath(site, "rest/networkconf"))
}

// GetFirewallRules fetches all firewall rules for the given site.
func (c *Client) GetFirewallRules(ctx context.Context, site string) ([]models.FirewallRule, error) {
	return fetch[models.FirewallRule](c, ctx, sitePath(site, "rest/firewallrule"))
}

// GetFirewallGroups fetches all firewall groups for the given site.
func (c *Client) GetFirewallGroups(ctx context.Context, site string) ([]models.FirewallGroup, error) {
	return fetch[models.FirewallGroup](c, ctx, sitePath(site, "rest/firewallgroup"))
}

// GetWLANs fetches all wireless network configurations for the given site.
func (c *Client) GetWLANs(ctx context.Context, site string) ([]models.WLAN, error) {
	return fetch[models.WLAN](c, ctx, sitePath(site, "rest/wlanconf"))
}

// GetPortForwards fetches all port forwarding rules for the given site.
func (c *Client) GetPortForwards(ctx context.Context, site string) ([]models.PortForward, error) {
	return fetch[models.PortForward](c, ctx, sitePath(site, "rest/portforward"))
}

// GetPortProfiles fetches all switch port profiles for the given site.
func (c *Client) GetPortProfiles(ctx context.Context, site string) ([]models.PortProfile, error) {
	return fetch[models.PortProfile](c, ctx, sitePath(site, "rest/portconf"))
}

// GetStaticRoutes fetches all static routes for the given site.
func (c *Client) GetStaticRoutes(ctx context.Context, site string) ([]models.StaticRoute, error) {
	return fetch[models.StaticRoute](c, ctx, sitePath(site, "rest/routing"))
}

// GetDevices fetches all device stats for the given site.
func (c *Client) GetDevices(ctx context.Context, site string) ([]models.Device, error) {
	return fetch[models.Device](c, ctx, sitePath(site, "stat/device"))
}
