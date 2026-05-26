// Package main is the entry point for the unifi-bootstrapper CLI tool.
// It connects to a UniFi UDM-Pro local API, enumerates Terraform-manageable
// resources, and emits JSON/Markdown inventory and Terraform import files.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/clcollins/unifi-bootstrapper/internal/client"
	"github.com/clcollins/unifi-bootstrapper/internal/exporter"
	"github.com/clcollins/unifi-bootstrapper/internal/generator"
	"github.com/clcollins/unifi-bootstrapper/internal/renderer"
)

// version is set at build time via -ldflags.
var version = "dev"

// testHTTPClient allows tests to inject an httptest-compatible client.
// In production this is always nil and the client creates its own.
var testHTTPClient *http.Client

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd creates the root Cobra command with all subcommands and
// persistent flags bound to Viper environment variables.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "unifi-bootstrapper",
		Short: "Export UniFi resources and generate Terraform import files",
		Long: `unifi-bootstrapper connects to a UniFi UDM-Pro local API,
enumerates Terraform-manageable resources, and emits JSON/Markdown
inventory files and Terraform import/stub configurations.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Persistent flags available to all subcommands.
	pflags := rootCmd.PersistentFlags()
	pflags.String("host", "", "UDM-Pro URL (env: UNIFI_HOST)")
	pflags.String("api-key", "", "API key for authentication (env: UNIFI_API_KEY)")
	pflags.String("username", "", "Username for cookie-session auth (env: UNIFI_USERNAME)")
	pflags.String("password", "", "Password for cookie-session auth (env: UNIFI_PASSWORD)")
	pflags.String("site", "default", "UniFi site name (env: UNIFI_SITE)")
	pflags.Bool("insecure", false, "Skip TLS certificate verification (env: UNIFI_INSECURE)")

	// Bind flags to Viper.
	viper.SetEnvPrefix("UNIFI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	_ = viper.BindPFlag("host", pflags.Lookup("host"))
	_ = viper.BindPFlag("api_key", pflags.Lookup("api-key"))
	_ = viper.BindPFlag("username", pflags.Lookup("username"))
	_ = viper.BindPFlag("password", pflags.Lookup("password"))
	_ = viper.BindPFlag("site", pflags.Lookup("site"))
	_ = viper.BindPFlag("insecure", pflags.Lookup("insecure"))

	// Add subcommands.
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newPingCmd())
	rootCmd.AddCommand(newExportCmd())

	return rootCmd
}

// newVersionCmd creates the version subcommand.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unifi-bootstrapper version %s\n", version)
			return nil
		},
	}
}

// newPingCmd creates the ping subcommand that tests API connectivity.
func newPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Test connectivity to the UniFi controller",
		RunE: func(cmd *cobra.Command, _ []string) error {
			host := viper.GetString("host")
			if host == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: --host is required (or set UNIFI_HOST)")
				return fmt.Errorf("host is required")
			}

			c := buildClient(host)
			ctx := context.Background()

			if err := c.Ping(ctx); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: ping failed: %v\n", err)
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Connection successful")
			return nil
		},
	}
}

// newExportCmd creates the export subcommand that fetches all resources
// and writes output files.
func newExportCmd() *cobra.Command {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export UniFi resources to JSON, Markdown, and Terraform files",
		RunE: func(cmd *cobra.Command, _ []string) error {
			host := viper.GetString("host")
			if host == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error: --host is required (or set UNIFI_HOST)")
				return fmt.Errorf("host is required")
			}

			outDir := viper.GetString("out_dir")
			site := viper.GetString("site")

			c := buildClient(host)
			exp := exporter.NewExporter(c, site)

			ctx := context.Background()
			inv, err := exp.Export(ctx)

			// On partial failure, print warnings but continue writing
			// whatever data was successfully retrieved.
			var partialFailure bool
			if err != nil {
				partialFailure = true
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: some endpoints failed: %v\n", err)
			}

			// Render outputs.
			jsonBytes, jsonErr := renderer.RenderJSON(inv)
			if jsonErr != nil {
				return fmt.Errorf("rendering JSON: %w", jsonErr)
			}
			mdContent := renderer.RenderMarkdown(inv)
			providerContent := generator.GenerateProvider()
			importsContent := generator.GenerateImports(inv)
			stubsContent := generator.GenerateStubs(inv)

			// Create output directories.
			terraformDir := filepath.Join(outDir, "terraform")
			if mkdirErr := os.MkdirAll(terraformDir, 0o755); mkdirErr != nil {
				return fmt.Errorf("creating output directories: %w", mkdirErr)
			}

			// Write files.
			files := map[string][]byte{
				filepath.Join(outDir, "inventory.json"):          jsonBytes,
				filepath.Join(outDir, "inventory.md"):            []byte(mdContent),
				filepath.Join(terraformDir, "provider.tf"):       []byte(providerContent),
				filepath.Join(terraformDir, "imports.tf"):        []byte(importsContent),
				filepath.Join(terraformDir, "stubs.tf"):          []byte(stubsContent),
			}

			for path, content := range files {
				if writeErr := os.WriteFile(path, content, 0o644); writeErr != nil {
					return fmt.Errorf("writing %s: %w", path, writeErr)
				}
			}

			// Print summary.
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, "Export complete.")
			_, _ = fmt.Fprintf(out, "  Networks:        %d\n", len(inv.Networks))
			_, _ = fmt.Fprintf(out, "  Firewall Rules:  %d\n", len(inv.FirewallRules))
			_, _ = fmt.Fprintf(out, "  Firewall Groups: %d\n", len(inv.FirewallGroups))
			_, _ = fmt.Fprintf(out, "  WLANs:           %d\n", len(inv.WLANs))
			_, _ = fmt.Fprintf(out, "  Port Forwards:   %d\n", len(inv.PortForwards))
			_, _ = fmt.Fprintf(out, "  Port Profiles:   %d\n", len(inv.PortProfiles))
			_, _ = fmt.Fprintf(out, "  Static Routes:   %d\n", len(inv.StaticRoutes))
			_, _ = fmt.Fprintf(out, "  Devices:         %d\n", len(inv.Devices))
			_, _ = fmt.Fprintln(out, "")
			_, _ = fmt.Fprintln(out, "Output files:")
			for path := range files {
				_, _ = fmt.Fprintf(out, "  %s\n", path)
			}

			if partialFailure {
				return fmt.Errorf("export completed with errors")
			}
			return nil
		},
	}

	exportCmd.Flags().String("out-dir", "output", "Output directory for exported files (env: UNIFI_OUT_DIR)")
	_ = viper.BindPFlag("out_dir", exportCmd.Flags().Lookup("out-dir"))

	return exportCmd
}

// buildClient creates a client.Client from the current Viper configuration.
func buildClient(host string) *client.Client {
	var opts []client.Option

	if apiKey := viper.GetString("api_key"); apiKey != "" {
		opts = append(opts, client.WithAPIKey(apiKey))
	}

	username := viper.GetString("username")
	password := viper.GetString("password")
	if username != "" && password != "" {
		opts = append(opts, client.WithCredentials(username, password))
	}

	if viper.GetBool("insecure") {
		opts = append(opts, client.WithInsecure(true))
	}

	if testHTTPClient != nil {
		opts = append(opts, client.WithHTTPClient(testHTTPClient))
	}

	return client.NewClient(host, opts...)
}
