package models

// PortForward represents a UniFi port forwarding rule as returned by
// the UDM-Pro API. Port forwards map external ports to internal hosts.
type PortForward struct {
	ID            string `json:"_id"`
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	Src           string `json:"src,omitempty"`
	DstPort       string `json:"dst_port"`
	Fwd           string `json:"fwd"`
	FwdPort       string `json:"fwd_port"`
	Proto         string `json:"proto"`
	Log           bool   `json:"log"`
	SiteID        string `json:"site_id"`
	PfwdInterface string `json:"pfwd_interface,omitempty"`
}
