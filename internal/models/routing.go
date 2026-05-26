package models

// StaticRoute represents a UniFi static route as returned by the
// UDM-Pro API. Static routes define custom routing entries for the
// gateway, including nexthop, interface, and blackhole routes.
type StaticRoute struct {
	ID                 string `json:"_id"`
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	Type               string `json:"type"`
	StaticRouteNetwork string `json:"static-route_network"`
	GatewayIP          string `json:"gateway_ip,omitempty"`
	GatewayType        string `json:"gateway_type,omitempty"`
	InterfaceName      string `json:"interface,omitempty"`
	Distance           int    `json:"distance,omitempty"`
	SiteID             string `json:"site_id"`
}
