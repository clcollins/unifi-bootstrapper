// Package renderer formats exported UniFi resource data into
// human-readable JSON and Markdown inventory files. All renderers
// are pure functions: they take an Inventory, return output, and
// have no side effects or file I/O.
//
// IMPORTANT: All renderers MUST redact sensitive data before
// producing output. WLAN passphrases (XPassphrase) are replaced
// with "REDACTED" in every output format.
package renderer
