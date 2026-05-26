// Package models defines the data structures representing UniFi resources
// such as networks, WLANs, firewall rules, port profiles, and devices.
package models

// APIResponse wraps the standard UDM-Pro API response envelope.
// The UDM-Pro API returns all responses in this format with a meta
// field indicating success/failure and a data array of the requested
// resource type.
type APIResponse[T any] struct {
	Meta struct {
		RC string `json:"rc"`
	} `json:"meta"`
	Data []T `json:"data"`
}
