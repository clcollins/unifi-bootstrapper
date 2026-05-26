package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown_HappyPath(t *testing.T) {
	inv := testInventory()

	output := RenderMarkdown(inv)
	assert.NotEmpty(t, output)

	// Verify main title
	assert.Contains(t, output, "# UniFi Network Inventory")

	// Verify export timestamp is present
	assert.Contains(t, output, "Exported at:")
}

func TestRenderMarkdown_PassphraseRedaction(t *testing.T) {
	inv := testInventory()

	output := RenderMarkdown(inv)

	// The output MUST contain REDACTED where passphrases were
	assert.Contains(t, output, "REDACTED",
		"Markdown output must contain REDACTED for passphrase fields")

	// The output MUST NOT contain ANY of the known passphrase values
	for _, passphrase := range knownPassphrases {
		assert.False(t, strings.Contains(output, passphrase),
			"Markdown output must NOT contain known passphrase %q", passphrase)
	}
}

func TestRenderMarkdown_EmptyInventory(t *testing.T) {
	inv := &models.Inventory{
		ExportedAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
	}

	output := RenderMarkdown(inv)
	assert.NotEmpty(t, output)

	// Headers must still be present even with empty data
	assert.Contains(t, output, "# UniFi Network Inventory")
	assert.Contains(t, output, "## Networks (0)")
	assert.Contains(t, output, "## WLANs (0)")
	assert.Contains(t, output, "## Devices (0)")
}

func TestRenderMarkdown_ExpectedSections(t *testing.T) {
	inv := testInventory()

	output := RenderMarkdown(inv)

	expectedSections := []string{
		"## Networks (",
		"## Firewall Rules (",
		"## Firewall Groups (",
		"## WLANs (",
		"## Port Forwards (",
		"## Port Profiles (",
		"## Static Routes (",
		"## Devices (",
	}
	for _, section := range expectedSections {
		assert.Contains(t, output, section,
			"Markdown output must contain section header %q", section)
	}
}

func TestRenderMarkdown_ResourceCounts(t *testing.T) {
	inv := testInventory()

	output := RenderMarkdown(inv)

	// Verify counts match the test inventory
	assert.Contains(t, output, "## Networks (2)")
	assert.Contains(t, output, "## Firewall Rules (1)")
	assert.Contains(t, output, "## Firewall Groups (1)")
	assert.Contains(t, output, "## WLANs (2)")
	assert.Contains(t, output, "## Port Forwards (1)")
	assert.Contains(t, output, "## Port Profiles (1)")
	assert.Contains(t, output, "## Static Routes (1)")
	assert.Contains(t, output, "## Devices (1)")
}

func TestRenderMarkdown_TableContent(t *testing.T) {
	inv := testInventory()

	output := RenderMarkdown(inv)

	// Verify specific data appears in the tables
	assert.Contains(t, output, "LAN")
	assert.Contains(t, output, "192.168.1.0/24")
	assert.Contains(t, output, "Block IoT to LAN")
	assert.Contains(t, output, "HomeNet-5G")
	assert.Contains(t, output, "UDM-Pro")
	assert.Contains(t, output, "aa:bb:cc:dd:ee:ff")
}

func TestRenderMarkdown_DoesNotMutateInput(t *testing.T) {
	inv := testInventory()
	originalPassphrase := inv.WLANs[0].XPassphrase

	_ = RenderMarkdown(inv)

	// The original inventory must NOT be modified
	assert.Equal(t, originalPassphrase, inv.WLANs[0].XPassphrase,
		"RenderMarkdown must not mutate the input inventory")
}
