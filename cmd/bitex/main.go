// Bitex CLI entry point - Forensic disk metadata extraction tool
package main

import (
	"fmt"
	"os"

	"github.com/ilexum-group/bitex/internal/acquisition"
	"github.com/ilexum-group/bitex/internal/config"
	"github.com/ilexum-group/bitex/internal/logger"
	internalos "github.com/ilexum-group/bitex/internal/os"
	"github.com/ilexum-group/bitex/internal/sender"
	"github.com/ilexum-group/bitex/internal/tsk"
	"github.com/ilexum-group/bitex/pkg/models"
)

const (
	applicationName = "Bitex"
)

// Set via -ldflags at build time.
var version = "placeholder"

func main() {
	// Parse and validate configuration
	cfg := parseAndValidateConfig()

	// Create custody chain entry
	custodyChainEntry := createCustodyChain()

	// Initialize OS abstraction layer
	osImpl := initializeOS()
	osImpl.SetLogger(custodyChainEntry.LogCommand)

	// Set agent information in custody chain
	hostname, err := osImpl.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	currentUser, err := osImpl.GetCurrentUser()
	if err != nil {
		currentUser = "unknown"
	}

	custodyChainEntry.SetAgentUser(currentUser)
	custodyChainEntry.SetAgentHostname(hostname)

	// Initialize logging system
	pid := osImpl.GetProcessID()
	initializeLogging(hostname, currentUser, pid)

	// Perform disk acquisition
	analysis := performAcquisition(osImpl, cfg, custodyChainEntry)
	analysis.CaseID = cfg.CaseID

	// Send analysis to server if configured
	if cfg.ServerURL != "" {
		sendToServer(cfg, analysis)
	} else {
		logger.LogInfo("AnalysisComplete", map[string]string{
			"message": "Analysis completed successfully (no server configured)",
			"files":   fmt.Sprintf("%d", len(analysis.FileListing)),
		})
	}

	logger.LogInfo("BitexComplete", map[string]string{"status": "success"})
}

// ============================================================================
// Initialization Functions
// ============================================================================

// parseAndValidateConfig parses command-line flags and validates the configuration.
// Exits with error code 1 if validation fails.
func parseAndValidateConfig() *config.Config {
	cfg := config.ParseFlags()

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	return cfg
}

// initializeOS creates and initializes the OS abstraction layer.
// Exits with error code 1 if initialization fails.
func initializeOS() internalos.OS {
	osImpl := internalos.New()

	// Verify OS initialization by checking hostname
	if _, err := osImpl.Hostname(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize OS abstraction: %v\n", err)
		os.Exit(1)
	}

	return osImpl
}

// initializeLogging sets up the logging system with OS information.
// Logs system information including hostname, user, and process ID.
func initializeLogging(hostname string, currentUser string, processID int) {
	// Initialize logger with system information
	if err := logger.InitDefaultLogger(applicationName, hostname, fmt.Sprintf("%d", processID)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.LogInfo("SystemInfo", map[string]string{
		"hostname": hostname,
		"user":     currentUser,
		"pid":      fmt.Sprintf("%d", processID),
		"version":  version,
	})
}

// createCustodyChain creates and initializes the custody chain entry.
// Exits with error code 1 if creation fails.
func createCustodyChain() *models.CustodyChainEntry {
	custodyChainEntry, err := models.NewCustodyChainEntry(
		applicationName,
		version,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create custody chain: %v\n", err)
		os.Exit(1)
	}

	logger.LogInfo("CustodyChainInitialized", map[string]string{
		"agent":   applicationName,
		"version": version,
	})

	return custodyChainEntry
}

// ============================================================================
// Core Operations
// ============================================================================

// performAcquisition executes the disk acquisition process.
// Exits with error code 1 if acquisition fails.
func performAcquisition(osImpl internalos.OS, cfg *config.Config, custodyChainEntry *models.CustodyChainEntry) *models.TSKAnalysis {
	logger.LogInfo("AcquisitionStart", map[string]string{
		"disk":    cfg.DiskPath,
		"caseID":  cfg.CaseID,
		"version": version,
	})

	// Initialize TSK analyzer
	tskAnalyzer := tsk.NewTSKAnalyzer(custodyChainEntry, &osImpl)

	// Create acquirer
	acquirer := acquisition.NewAcquirer(osImpl, cfg.DiskPath, custodyChainEntry, tskAnalyzer)

	// Perform disk acquisition
	disk, err := acquirer.AcquireDisk()
	if err != nil {
		logger.LogError("AcquisitionFailed", map[string]string{"error": err.Error()})
		os.Exit(1)
	}

	// Finalize analysis with custody chain
	analysis := acquirer.GetAnalysisWithCustody(disk)

	// Set case ID if provided
	if cfg.CaseID != "" {
		analysis.CaseID = cfg.CaseID
	}

	logger.LogInfo("AcquisitionComplete", map[string]string{
		"files":  fmt.Sprintf("%d", len(analysis.FileListing)),
		"size":   fmt.Sprintf("%d", analysis.CustodyChain.TotalSizeBytes),
		"caseID": cfg.CaseID,
	})

	return analysis
}

// sendToServer transmits the analysis to the configured server.
// Exits with error code 1 if transmission fails.
func sendToServer(cfg *config.Config, analysis *models.TSKAnalysis) {
	logger.LogInfo("TransmissionStart", map[string]string{
		"server": cfg.ServerURL,
		"files":  fmt.Sprintf("%d", len(analysis.FileListing)),
	})

	sender := sender.NewSender(cfg.ServerURL, cfg.AuthToken)
	if err := sender.SendAnalysis(analysis); err != nil {
		logger.LogError("TransmissionFailed", map[string]string{"error": err.Error()})
		os.Exit(1)
	}

	logger.LogInfo("TransmissionComplete", map[string]string{
		"server": cfg.ServerURL,
		"status": "success",
	})
}
