package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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

// fixtureHandler returns an http.Handler that serves fixture data for
// all UDM-Pro API endpoints, suitable for integration testing with a
// real client. The health endpoint returns empty success for Ping.
func fixtureHandler(t *testing.T) http.Handler {
	t.Helper()
	fixtures := map[string]string{
		"/proxy/network/api/s/default/rest/networkconf":   "networks.json",
		"/proxy/network/api/s/default/rest/firewallrule":  "firewall_rules.json",
		"/proxy/network/api/s/default/rest/firewallgroup": "firewall_groups.json",
		"/proxy/network/api/s/default/rest/wlanconf":      "wlans.json",
		"/proxy/network/api/s/default/rest/portforward":   "port_forwards.json",
		"/proxy/network/api/s/default/rest/portconf":      "port_profiles.json",
		"/proxy/network/api/s/default/rest/routing":       "static_routes.json",
		"/proxy/network/api/s/default/stat/device":        "devices.json",
		"/proxy/network/api/s/default/stat/health":        "",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixtureName, ok := fixtures[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"meta":{"rc":"error","msg":"not found"}}`))
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
}

// authFailureHandler returns an http.Handler that responds with 401 to
// every request, simulating an authentication failure.
func authFailureHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"meta":{"rc":"error","msg":"api.err.Unauthorized"},"data":[]}`))
	})
}

// partialFailureHandler returns an http.Handler that serves fixture data
// for most endpoints but returns a server error for one endpoint
// (firewall rules), simulating partial export failure.
func partialFailureHandler(t *testing.T) http.Handler {
	t.Helper()
	fixtures := map[string]string{
		"/proxy/network/api/s/default/rest/networkconf":   "networks.json",
		"/proxy/network/api/s/default/rest/firewallgroup": "firewall_groups.json",
		"/proxy/network/api/s/default/rest/wlanconf":      "wlans.json",
		"/proxy/network/api/s/default/rest/portforward":   "port_forwards.json",
		"/proxy/network/api/s/default/rest/portconf":      "port_profiles.json",
		"/proxy/network/api/s/default/rest/routing":       "static_routes.json",
		"/proxy/network/api/s/default/stat/device":        "devices.json",
		"/proxy/network/api/s/default/stat/health":        "",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Firewall rules endpoint returns an error
		if r.URL.Path == "/proxy/network/api/s/default/rest/firewallrule" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`Internal Server Error`))
			return
		}

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
}

// --- version command tests ---

func TestVersionCommand_PrintsVersion(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "unifi-bootstrapper version")
	assert.Contains(t, stdout.String(), "dev")
}

// --- ping command tests ---

func TestPingCommand_Success(t *testing.T) {
	server := httptest.NewTLSServer(fixtureHandler(t))
	t.Cleanup(server.Close)

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"ping",
		"--host", server.URL,
		"--api-key", "fake-api-key-for-test",
		"--insecure",
	})

	// Override the HTTP client for the test server
	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Connection successful")
}

func TestPingCommand_AuthFailure(t *testing.T) {
	server := httptest.NewTLSServer(authFailureHandler())
	t.Cleanup(server.Close)

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"ping",
		"--host", server.URL,
		"--api-key", "bad-api-key",
		"--insecure",
	})

	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "401")
}

func TestPingCommand_MissingHost(t *testing.T) {
	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"ping"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "host")
}

// --- export command tests ---

func TestExportCommand_Success(t *testing.T) {
	server := httptest.NewTLSServer(fixtureHandler(t))
	t.Cleanup(server.Close)

	outDir := t.TempDir()

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"export",
		"--host", server.URL,
		"--api-key", "fake-api-key-for-test",
		"--insecure",
		"--out-dir", outDir,
	})

	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify all 5 output files are created
	expectedFiles := []string{
		filepath.Join(outDir, "inventory.json"),
		filepath.Join(outDir, "inventory.md"),
		filepath.Join(outDir, "terraform", "provider.tf"),
		filepath.Join(outDir, "terraform", "imports.tf"),
		filepath.Join(outDir, "terraform", "stubs.tf"),
	}

	for _, path := range expectedFiles {
		info, err := os.Stat(path)
		require.NoError(t, err, "expected file to exist: %s", path)
		assert.Greater(t, info.Size(), int64(0), "file should be non-empty: %s", path)
	}

	// Verify inventory.json is valid JSON
	jsonData, err := os.ReadFile(filepath.Join(outDir, "inventory.json"))
	require.NoError(t, err)
	assert.True(t, json.Valid(jsonData), "inventory.json should be valid JSON")

	// Verify inventory.md has expected header
	mdData, err := os.ReadFile(filepath.Join(outDir, "inventory.md"))
	require.NoError(t, err)
	assert.Contains(t, string(mdData), "# UniFi Network Inventory")

	// Verify terraform/provider.tf has provider block
	providerData, err := os.ReadFile(filepath.Join(outDir, "terraform", "provider.tf"))
	require.NoError(t, err)
	assert.Contains(t, string(providerData), "filipowm/unifi")

	// Verify summary output mentions resource counts
	assert.Contains(t, stdout.String(), "Networks:")
	assert.Contains(t, stdout.String(), "Devices:")
}

func TestExportCommand_PartialFailure(t *testing.T) {
	server := httptest.NewTLSServer(partialFailureHandler(t))
	t.Cleanup(server.Close)

	outDir := t.TempDir()

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"export",
		"--host", server.URL,
		"--api-key", "fake-api-key-for-test",
		"--insecure",
		"--out-dir", outDir,
	})

	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	err := cmd.Execute()
	// Partial failure should cause the command to return an error
	require.Error(t, err)

	// Output files should still be created with partial data
	expectedFiles := []string{
		filepath.Join(outDir, "inventory.json"),
		filepath.Join(outDir, "inventory.md"),
		filepath.Join(outDir, "terraform", "provider.tf"),
		filepath.Join(outDir, "terraform", "imports.tf"),
		filepath.Join(outDir, "terraform", "stubs.tf"),
	}

	for _, path := range expectedFiles {
		_, statErr := os.Stat(path)
		assert.NoError(t, statErr, "expected file to exist even on partial failure: %s", path)
	}

	// Warning should be printed about the failed endpoint
	assert.Contains(t, stderr.String(), "warning", "should print warning about failed endpoint")
}

func TestExportCommand_MissingHost(t *testing.T) {
	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"export",
		"--out-dir", t.TempDir(),
	})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "host")
}

// --- help output tests ---

func TestRootHelp_ShowsSubcommands(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "ping")
	assert.Contains(t, output, "export")
	assert.Contains(t, output, "version")
}

func TestPingHelp_ShowsFlags(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"ping", "--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "--host")
	assert.Contains(t, output, "--api-key")
	assert.Contains(t, output, "--insecure")
}

func TestExportHelp_ShowsFlags(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"export", "--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "--host")
	assert.Contains(t, output, "--out-dir")
	assert.Contains(t, output, "--api-key")
}

// --- Viper env var binding tests ---

func TestViperEnvBinding_Host(t *testing.T) {
	// This test verifies that the UNIFI_HOST env var is picked up
	// by Viper when the --host flag is not set.
	server := httptest.NewTLSServer(fixtureHandler(t))
	t.Cleanup(server.Close)

	t.Setenv("UNIFI_HOST", server.URL)
	t.Setenv("UNIFI_API_KEY", "fake-env-api-key")
	t.Setenv("UNIFI_INSECURE", "true")

	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"ping"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Connection successful")
}

// --- export creates directories ---

func TestExportCommand_CreatesDirectories(t *testing.T) {
	server := httptest.NewTLSServer(fixtureHandler(t))
	t.Cleanup(server.Close)

	// Use a nested path that does not exist yet
	outDir := filepath.Join(t.TempDir(), "nested", "output")

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"export",
		"--host", server.URL,
		"--api-key", "fake-api-key-for-test",
		"--insecure",
		"--out-dir", outDir,
	})

	testHTTPClient = server.Client()
	t.Cleanup(func() { testHTTPClient = nil })

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify the nested directory was created
	_, err = os.Stat(filepath.Join(outDir, "terraform", "provider.tf"))
	require.NoError(t, err, "nested directories should be created by export")
}
