package models

import "time"

// Inventory aggregates all UniFi resource types into a single
// structure for export and rendering. The ExportedAt field records
// when the inventory snapshot was taken.
type Inventory struct {
	Networks       []Network       `json:"networks"`
	FirewallRules  []FirewallRule  `json:"firewall_rules"`
	FirewallGroups []FirewallGroup `json:"firewall_groups"`
	WLANs          []WLAN          `json:"wlans"`
	PortForwards   []PortForward   `json:"port_forwards"`
	PortProfiles   []PortProfile   `json:"port_profiles"`
	StaticRoutes   []StaticRoute   `json:"static_routes"`
	Devices        []Device        `json:"devices"`
	ExportedAt     time.Time       `json:"exported_at"`
}
