package models

// PortProfile represents a UniFi switch port profile as returned by
// the UDM-Pro API. Port profiles define VLAN assignment, PoE settings,
// speed/duplex, and storm control for switch ports.
type PortProfile struct {
	ID                    string   `json:"_id"`
	Name                  string   `json:"name"`
	NativeNetworkconfID   string   `json:"native_networkconf_id,omitempty"`
	TaggedNetworkconfIDs  []string `json:"tagged_networkconf_ids"`
	PoeMode               string   `json:"poe_mode,omitempty"`
	Forward               string   `json:"forward,omitempty"`
	SiteID                string   `json:"site_id"`
	Autoneg               bool     `json:"autoneg"`
	Speed                 int      `json:"speed,omitempty"`
	FullDuplex            bool     `json:"full_duplex"`
	StormctrlBcastEnabled bool     `json:"stormctrl_bcast_enabled"`
	StormctrlMcastEnabled bool     `json:"stormctrl_mcast_enabled"`
	StormctrlUcastEnabled bool     `json:"stormctrl_ucast_enabled"`
}
