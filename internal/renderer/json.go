package renderer

import (
	"encoding/json"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

// redactedValue is the placeholder string that replaces sensitive
// data in rendered output.
const redactedValue = "REDACTED"

// RenderJSON takes an Inventory and returns JSON bytes with all
// sensitive data redacted. It creates a deep copy of the WLANs
// slice so the original Inventory is not mutated.
//
// Output uses 2-space indentation for human readability.
// This is a pure function — no file I/O or side effects.
func RenderJSON(inv *models.Inventory) ([]byte, error) {
	redacted := redactInventory(inv)
	return json.MarshalIndent(redacted, "", "  ")
}

// redactInventory returns a shallow copy of the Inventory with all
// sensitive fields replaced. Currently redacts WLAN XPassphrase
// values. The original Inventory is not modified.
func redactInventory(inv *models.Inventory) models.Inventory {
	out := *inv

	// Deep copy and redact WLANs
	if len(inv.WLANs) > 0 {
		out.WLANs = make([]models.WLAN, len(inv.WLANs))
		copy(out.WLANs, inv.WLANs)
		for i := range out.WLANs {
			if out.WLANs[i].XPassphrase != "" {
				out.WLANs[i].XPassphrase = redactedValue
			}
		}
	}

	return out
}
