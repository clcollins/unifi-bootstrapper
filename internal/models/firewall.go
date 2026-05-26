package models

// FirewallRule represents a UniFi firewall rule as returned by the
// UDM-Pro API. Rules belong to a ruleset (e.g., LAN_IN, GUEST_IN) and
// define traffic matching criteria and actions.
//
// The RuleIndex field uses FlexInt because the UDM-Pro API returns it
// as a string on some firmware versions and as an integer on others.
type FirewallRule struct {
	ID              string  `json:"_id"`
	Name            string  `json:"name"`
	Action          string  `json:"action"`
	Enabled         bool    `json:"enabled"`
	RuleIndex       FlexInt `json:"rule_index"`
	Ruleset         string  `json:"ruleset"`
	Protocol        string  `json:"protocol,omitempty"`
	SrcAddress      string  `json:"src_address,omitempty"`
	DstAddress      string  `json:"dst_address,omitempty"`
	SrcPort         string  `json:"src_port,omitempty"`
	DstPort         string  `json:"dst_port,omitempty"`
	SrcNetworkID    string  `json:"src_networkconf_id,omitempty"`
	DstNetworkID    string  `json:"dst_networkconf_id,omitempty"`
	SrcNetworkType  string  `json:"src_networkconf_type,omitempty"`
	DstNetworkType  string  `json:"dst_networkconf_type,omitempty"`
	ICMPTypename    string  `json:"icmp_typename,omitempty"`
	StateNew        bool    `json:"state_new"`
	StateEstablished bool   `json:"state_established"`
	StateInvalid    bool    `json:"state_invalid"`
	StateRelated    bool    `json:"state_related"`
	Logging         bool    `json:"logging"`
	SiteID          string  `json:"site_id"`
}

// FirewallGroup represents a UniFi firewall group (address-group or
// port-group) as returned by the UDM-Pro API. Groups are referenced
// by firewall rules to match traffic against sets of addresses or ports.
type FirewallGroup struct {
	ID           string   `json:"_id"`
	Name         string   `json:"name"`
	GroupType    string   `json:"group_type"`
	GroupMembers []string `json:"group_members"`
	SiteID       string   `json:"site_id"`
}
