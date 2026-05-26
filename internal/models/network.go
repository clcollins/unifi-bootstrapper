package models

// Network represents a UniFi network configuration as returned by the
// UDM-Pro API. Networks can serve different purposes (corporate, guest,
// WAN, VLAN-only) and control DHCP, VLAN, and other layer-2/3 settings.
type Network struct {
	ID             string `json:"_id"`
	Name           string `json:"name"`
	Purpose        string `json:"purpose"`
	Subnet         string `json:"subnet,omitempty"`
	VLANID         int    `json:"vlan,omitempty"`
	VLANEnabled    bool   `json:"vlan_enabled"`
	DHCPEnabled    bool   `json:"dhcp_enabled"`
	DHCPStart      string `json:"dhcpd_start,omitempty"`
	DHCPStop       string `json:"dhcpd_stop,omitempty"`
	DHCPDNSEnabled bool   `json:"dhcpd_dns_enabled"`
	DomainName     string `json:"domain_name,omitempty"`
	IGMPSnooping   bool   `json:"igmp_snooping"`
	NetworkGroup   string `json:"networkgroup,omitempty"`
	SiteID         string `json:"site_id"`
}
