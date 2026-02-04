// Package config provides CLI configuration parsing and validation for Bitex.
package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds CLI configuration for Bitex.
type Config struct {
	DiskPath  string
	CaseID    string
	ServerURL string
	AuthToken string
}

const usage = `Bitex - Forensic Disk Analysis Binary
A portable, auditable forensic tool for disk image metadata extraction with chain of custody

Usage:
  bitex --disk <disk-image> --case-id <case-id> --server <url> --token <token>

Options:
  -h, --help             Show this help message
  -v, --version          Show version information
  --disk PATH            Path to disk image or block device (required)
  --case-id ID           Case identifier for correlation (required)
  --server URL           Remote server endpoint URL (required)
  --token TOKEN          Authentication token for remote server (required)

Examples:
  # Analyze and upload to server
  bitex --disk image.dd --case-id "CASE-2025-001" --server https://server.com/api/analysis --token TOKEN

  # Full example with block device
  bitex --disk /dev/sda --case-id "CASE-2025-001" --server https://server.com/api/analysis --token TOKEN

Forensic Analysis:
  All disk analysis is performed in read-only mode without any modifications.
  Uses The Sleuth Kit (TSK) for metadata-only extraction (no file content reading).
  Extracts file listing, timestamps, and deletion information.
  Every operation is logged and included in the forensic custody chain.
  Evidence is transmitted directly to the server without local storage.
`

// ParseFlags parses command-line flags for Bitex.
func ParseFlags() *Config {
	cfg := &Config{}

	fs := flag.NewFlagSet("bitex", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	fs.StringVar(&cfg.DiskPath, "disk", "", "Path to disk image or block device")
	fs.StringVar(&cfg.CaseID, "case-id", "", "Case identifier for correlation")
	fs.StringVar(&cfg.ServerURL, "server", "", "Remote server endpoint URL")
	fs.StringVar(&cfg.AuthToken, "token", "", "Authentication token for remote server")

	// Handle help and version
	helpFlag := fs.Bool("h", false, "Show help message")
	fs.Bool("help", false, "Show help message")
	versionFlag := fs.Bool("v", false, "Show version")
	fs.Bool("version", false, "Show version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("Bitex v1.0.0\n")
		os.Exit(0)
	}

	return cfg
}

// ValidateConfig validates the Bitex configuration.
func ValidateConfig(cfg *Config) error {
	// Disk path is required
	if cfg.DiskPath == "" {
		return fmt.Errorf("disk path (--disk) is required for analysis")
	}

	// Case ID is required
	if cfg.CaseID == "" {
		return fmt.Errorf("case identifier (--case-id) is required for correlation")
	}

	// Server URL is required
	if cfg.ServerURL == "" {
		return fmt.Errorf("server URL (--server) is required for results transmission")
	}

	// Authentication token is required
	if cfg.AuthToken == "" {
		return fmt.Errorf("authentication token (--token) is required for server communication")
	}

	// Validate disk path exists
	if _, err := os.Stat(cfg.DiskPath); err != nil {
		return fmt.Errorf("disk not found: %s", cfg.DiskPath)
	}

	return nil
}
