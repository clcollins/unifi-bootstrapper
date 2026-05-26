// Package exporter orchestrates concurrent fetches of all UniFi resource
// types from the UDM-Pro API via ClientInterface, aggregating results
// into a models.Inventory.
package exporter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/clcollins/unifi-bootstrapper/internal/client"
	"github.com/clcollins/unifi-bootstrapper/internal/models"
)

// Exporter fetches all resource types concurrently from a UniFi
// controller and assembles them into an Inventory.
type Exporter struct {
	client client.ClientInterface
	site   string
}

// NewExporter creates a new Exporter that will use the given client
// to fetch resources from the specified site.
func NewExporter(c client.ClientInterface, site string) *Exporter {
	return &Exporter{
		client: c,
		site:   site,
	}
}

// Export fetches all 8 resource types concurrently and returns an
// Inventory. On partial failure, it returns both the partial Inventory
// (containing successfully fetched resources) and a combined error
// describing all failures. On complete success, the error is nil.
func (e *Exporter) Export(ctx context.Context) (*models.Inventory, error) {
	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		inv    models.Inventory
		errs   []error
	)

	// fetchAndStore is a helper that runs a fetch function in a
	// goroutine and stores the result under the mutex.
	type storeFunc func()

	// Each fetch is dispatched as a goroutine. On success, the
	// result is written to the appropriate Inventory field. On
	// failure, the error is collected.
	fetchers := []struct {
		name  string
		fetch func() (storeFunc, error)
	}{
		{
			name: "networks",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetNetworks(ctx, e.site)
				return func() { inv.Networks = data }, err
			},
		},
		{
			name: "firewall_rules",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetFirewallRules(ctx, e.site)
				return func() { inv.FirewallRules = data }, err
			},
		},
		{
			name: "firewall_groups",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetFirewallGroups(ctx, e.site)
				return func() { inv.FirewallGroups = data }, err
			},
		},
		{
			name: "wlans",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetWLANs(ctx, e.site)
				return func() { inv.WLANs = data }, err
			},
		},
		{
			name: "port_forwards",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetPortForwards(ctx, e.site)
				return func() { inv.PortForwards = data }, err
			},
		},
		{
			name: "port_profiles",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetPortProfiles(ctx, e.site)
				return func() { inv.PortProfiles = data }, err
			},
		},
		{
			name: "static_routes",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetStaticRoutes(ctx, e.site)
				return func() { inv.StaticRoutes = data }, err
			},
		},
		{
			name: "devices",
			fetch: func() (storeFunc, error) {
				data, err := e.client.GetDevices(ctx, e.site)
				return func() { inv.Devices = data }, err
			},
		},
	}

	wg.Add(len(fetchers))

	for _, f := range fetchers {
		go func(name string, fetch func() (storeFunc, error)) {
			defer wg.Done()

			store, err := fetch()

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("fetching %s: %w", name, err))
				return
			}
			store()
		}(f.name, f.fetch)
	}

	wg.Wait()

	inv.ExportedAt = time.Now()

	return &inv, errors.Join(errs...)
}
