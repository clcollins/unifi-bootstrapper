package models

// Device represents a UniFi device (gateway, switch, access point) as
// returned by the UDM-Pro API. Devices are inventory-only — they are
// not managed via Terraform import but are included in the inventory
// for completeness.
type Device struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	ModelName string `json:"model_name,omitempty"`
	Type      string `json:"type"`
	Mac       string `json:"mac"`
	IP        string `json:"ip"`
	Version   string `json:"version,omitempty"`
	Serial    string `json:"serial,omitempty"`
	Adopted   bool   `json:"adopted"`
	State     int    `json:"state"`
	SiteID    string `json:"site_id"`
	Uptime    int    `json:"uptime,omitempty"`
}
