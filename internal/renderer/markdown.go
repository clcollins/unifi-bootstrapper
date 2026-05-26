package renderer

import (
	"fmt"
	"strings"

	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

// RenderMarkdown takes an Inventory and returns a Markdown string
// with all sensitive data redacted. The output contains sections
// for each resource type rendered as Markdown tables with key
// fields as columns. Section headers include resource counts.
//
// This is a pure function — no file I/O or side effects.
func RenderMarkdown(inv *models.Inventory) string {
	redacted := redactInventory(inv)
	var b strings.Builder

	b.WriteString("# UniFi Network Inventory\n\n")
	fmt.Fprintf(&b, "Exported at: %s\n", redacted.ExportedAt.Format("2006-01-02T15:04:05Z07:00"))

	renderNetworks(&b, redacted.Networks)
	renderFirewallRules(&b, redacted.FirewallRules)
	renderFirewallGroups(&b, redacted.FirewallGroups)
	renderWLANs(&b, redacted.WLANs)
	renderPortForwards(&b, redacted.PortForwards)
	renderPortProfiles(&b, redacted.PortProfiles)
	renderStaticRoutes(&b, redacted.StaticRoutes)
	renderDevices(&b, redacted.Devices)

	return b.String()
}

func renderNetworks(b *strings.Builder, networks []models.Network) {
	fmt.Fprintf(b, "\n## Networks (%d)\n\n", len(networks))
	b.WriteString("| Name | ID | Purpose | Subnet | VLAN |\n")
	b.WriteString("|------|-----|---------|--------|------|\n")
	for _, n := range networks {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %d |\n",
			n.Name, n.ID, n.Purpose, n.Subnet, n.VLANID)
	}
}

func renderFirewallRules(b *strings.Builder, rules []models.FirewallRule) {
	fmt.Fprintf(b, "\n## Firewall Rules (%d)\n\n", len(rules))
	b.WriteString("| Name | ID | Action | Ruleset | Protocol | Enabled |\n")
	b.WriteString("|------|-----|--------|---------|----------|---------|\n")
	for _, r := range rules {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %t |\n",
			r.Name, r.ID, r.Action, r.Ruleset, r.Protocol, r.Enabled)
	}
}

func renderFirewallGroups(b *strings.Builder, groups []models.FirewallGroup) {
	fmt.Fprintf(b, "\n## Firewall Groups (%d)\n\n", len(groups))
	b.WriteString("| Name | ID | Type | Members |\n")
	b.WriteString("|------|-----|------|----------|\n")
	for _, g := range groups {
		members := strings.Join(g.GroupMembers, ", ")
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n",
			g.Name, g.ID, g.GroupType, members)
	}
}

func renderWLANs(b *strings.Builder, wlans []models.WLAN) {
	fmt.Fprintf(b, "\n## WLANs (%d)\n\n", len(wlans))
	b.WriteString("| Name | ID | Security | Passphrase | VLAN | Band |\n")
	b.WriteString("|------|-----|----------|------------|------|------|\n")
	for _, w := range wlans {
		passphrase := w.XPassphrase
		if passphrase == "" {
			passphrase = "N/A"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s | %d | %s |\n",
			w.Name, w.ID, w.Security, passphrase, w.VLANID, w.WlanBand)
	}
}

func renderPortForwards(b *strings.Builder, forwards []models.PortForward) {
	fmt.Fprintf(b, "\n## Port Forwards (%d)\n\n", len(forwards))
	b.WriteString("| Name | ID | Dst Port | Forward To | Fwd Port | Proto | Enabled |\n")
	b.WriteString("|------|-----|----------|------------|----------|-------|---------|\n")
	for _, p := range forwards {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %t |\n",
			p.Name, p.ID, p.DstPort, p.Fwd, p.FwdPort, p.Proto, p.Enabled)
	}
}

func renderPortProfiles(b *strings.Builder, profiles []models.PortProfile) {
	fmt.Fprintf(b, "\n## Port Profiles (%d)\n\n", len(profiles))
	b.WriteString("| Name | ID | Forward | PoE Mode | Autoneg |\n")
	b.WriteString("|------|-----|---------|----------|---------|\n")
	for _, p := range profiles {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %t |\n",
			p.Name, p.ID, p.Forward, p.PoeMode, p.Autoneg)
	}
}

func renderStaticRoutes(b *strings.Builder, routes []models.StaticRoute) {
	fmt.Fprintf(b, "\n## Static Routes (%d)\n\n", len(routes))
	b.WriteString("| Name | ID | Type | Network | Gateway | Enabled |\n")
	b.WriteString("|------|-----|------|---------|---------|----------|\n")
	for _, r := range routes {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %t |\n",
			r.Name, r.ID, r.Type, r.StaticRouteNetwork, r.GatewayIP, r.Enabled)
	}
}

func renderDevices(b *strings.Builder, devices []models.Device) {
	fmt.Fprintf(b, "\n## Devices (%d)\n\n", len(devices))
	b.WriteString("| Name | ID | Model | IP | MAC | Version | State |\n")
	b.WriteString("|------|-----|-------|----|-----|---------|-------|\n")
	for _, d := range devices {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %d |\n",
			d.Name, d.ID, d.Model, d.IP, d.Mac, d.Version, d.State)
	}
}
