package models

// WLAN represents a UniFi wireless network as returned by the UDM-Pro
// API. WLANs define SSIDs with security, VLAN assignment, and band
// configuration.
//
// The XPassphrase field contains the WPA passphrase. It is included in
// the model for completeness and test coverage. Renderers and generators
// in later tasks are responsible for redacting this value in output.
type WLAN struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Security    string `json:"security"`
	WPAMode     string `json:"wpa_mode,omitempty"`
	XPassphrase string `json:"x_passphrase,omitempty"`
	VLANID      int    `json:"vlan,omitempty"`
	VLANEnabled bool   `json:"vlan_enabled"`
	IsGuest     bool   `json:"is_guest"`
	HideSsid    bool   `json:"hide_ssid"`
	WlanBand    string `json:"wlan_band,omitempty"`
	NetworkID   string `json:"networkconf_id,omitempty"`
	UserGroupID string `json:"usergroup_id,omitempty"`
	SiteID      string `json:"site_id"`
}
