// Bitex CLI entry point
package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/ilexum-group/bitex/internal/disk"
	"github.com/ilexum-group/bitex/internal/logger"
	"github.com/ilexum-group/bitex/internal/sender"
)

type cliConfig struct {
	DiskPath   string
	ServerURL  string
	AgentToken string
	CaseID     string
}

func main() {
	cfg := parseFlags()
	if cfg.DiskPath == "" {
		logger.Error("Error: --disk flag is required", nil)
		os.Exit(1)
	}

	logger.Info("Starting Bitex analysis", map[string]string{"disk": cfg.DiskPath})

	analysis, err := disk.AnalyzeDisk(cfg.DiskPath)
	if err != nil {
		logger.Error("Bitex error", map[string]string{"error": err.Error()})
		os.Exit(2)
	}

	if cfg.CaseID != "" {
		analysis.CaseID = cfg.CaseID
	}

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal output", map[string]string{"error": err.Error()})
		os.Exit(3)
	}

	logger.Info("Analysis result", map[string]string{"output": string(jsonBytes)})

	if cfg.ServerURL != "" {
		sendCfg := &sender.Config{
			ServerURL:  cfg.ServerURL,
			AgentToken: cfg.AgentToken,
		}
		if err := sender.SendAnalysis(sendCfg, analysis); err != nil {
			logger.Error("Failed to send analysis", map[string]string{"error": err.Error()})
			os.Exit(4)
		}
	}
}

func parseFlags() cliConfig {
	diskPath := flag.String("disk", "", "Path to disk image or block device (read-only)")
	serverURL := flag.String("server-url", "", "URL to send analysis results (optional)")
	agentToken := flag.String("agent-token", "", "Bearer token for authentication (optional)")
	caseID := flag.String("case-id", "", "Case identifier for correlation (optional)")
	flag.Parse()

	return cliConfig{
		DiskPath:   *diskPath,
		ServerURL:  *serverURL,
		AgentToken: *agentToken,
		CaseID:     *caseID,
	}
}
