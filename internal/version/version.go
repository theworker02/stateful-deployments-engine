// Package version is the single authoritative product version.
package version

// Version is the SemVer release string for CLI, receipts, SBOM, and docs.
const Version = "1.1.0"

// Codename is a human-facing release designation.
const Codename = "READY"

// String returns "sde <version>".
func String() string {
	return "sde " + Version
}
