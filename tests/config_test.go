package tests

import (
	"flag"
	"os"
	"testing"

	"github.com/ilexum-group/bitex/internal/config"
)

func TestParseFlags(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flag.CommandLine for each test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	tests := []struct {
		name     string
		args     []string
		expected config.Config
	}{
		{
			name: "All flags provided",
			args: []string{"cmd", "--disk", "/dev/sda", "--case-id", "CASE-001", "--server", "https://server.com", "--token", "secret"},
			expected: config.Config{
				DiskPath:  "/dev/sda",
				CaseID:    "CASE-001",
				ServerURL: "https://server.com",
				AuthToken: "secret",
			},
		},
		{
			name: "Only disk path provided",
			args: []string{"cmd", "--disk", "image.dd"},
			expected: config.Config{
				DiskPath:  "image.dd",
				CaseID:    "",
				ServerURL: "",
				AuthToken: "",
			},
		},
		{
			name: "Disk and case ID only",
			args: []string{"cmd", "--disk", "/path/to/disk.img", "--case-id", "TEST-123"},
			expected: config.Config{
				DiskPath:  "/path/to/disk.img",
				CaseID:    "TEST-123",
				ServerURL: "",
				AuthToken: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag.CommandLine for each subtest
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args
			cfg := config.ParseFlags()

			if cfg.DiskPath != tt.expected.DiskPath {
				t.Errorf("DiskPath = %v, want %v", cfg.DiskPath, tt.expected.DiskPath)
			}
			if cfg.CaseID != tt.expected.CaseID {
				t.Errorf("CaseID = %v, want %v", cfg.CaseID, tt.expected.CaseID)
			}
			if cfg.ServerURL != tt.expected.ServerURL {
				t.Errorf("ServerURL = %v, want %v", cfg.ServerURL, tt.expected.ServerURL)
			}
			if cfg.AuthToken != tt.expected.AuthToken {
				t.Errorf("AuthToken = %v, want %v", cfg.AuthToken, tt.expected.AuthToken)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "Valid config - disk only",
			config: &config.Config{
				DiskPath:  "testdata/test.dd",
				CaseID:    "CASE-001",
				ServerURL: "",
				AuthToken: "",
			},
			wantErr: true, // Server URL is required
		},
		{
			name: "Valid config - with server and token",
			config: &config.Config{
				DiskPath:  "testdata/test.dd",
				CaseID:    "CASE-001",
				ServerURL: "https://server.com",
				AuthToken: "token123",
			},
			wantErr: false,
		},
		{
			name: "Invalid - missing disk path",
			config: &config.Config{
				DiskPath:  "",
				CaseID:    "CASE-001",
				ServerURL: "",
				AuthToken: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid - disk path does not exist",
			config: &config.Config{
				DiskPath:  "/nonexistent/path/to/disk.dd",
				CaseID:    "",
				ServerURL: "",
				AuthToken: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid - server without token",
			config: &config.Config{
				DiskPath:  "testdata/test.dd",
				CaseID:    "",
				ServerURL: "https://server.com",
				AuthToken: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid - token without server",
			config: &config.Config{
				DiskPath:  "testdata/test.dd",
				CaseID:    "CASE-001",
				ServerURL: "",
				AuthToken: "token123",
			},
			wantErr: true, // Server URL is required
		},
	}

	// Create testdata directory and file for tests
	if err := os.MkdirAll("testdata", 0750); err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}
	testFile, err := os.Create("testdata/test.dd")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := testFile.Close(); err != nil {
		t.Logf("Warning: failed to close test file: %v", err)
	}
	defer func() {
		if err := os.RemoveAll("testdata"); err != nil {
			t.Logf("Warning: failed to remove testdata: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
