// Package sender handles sending analysis results to a remote server
package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
       
	"net/http"

	"github.com/ilexum-group/bitex/internal/logger"
)

type Config struct {
	ServerURL  string
	AgentToken string
}

// SendAnalysis sends the analysis result to the server as JSON
func SendAnalysis(cfg *Config, analysis interface{}) error {
	logger.Info("Preparing to send analysis to server", map[string]string{"url": cfg.ServerURL})

	jsonData, err := json.Marshal(analysis)
	if err != nil {
		logger.Error("Failed to marshal analysis", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	contentLength := len(jsonData)
	logger.Debug("Sending JSON payload", map[string]string{
		"content_length": fmt.Sprintf("%d bytes", contentLength),
	})

	req, err := http.NewRequest("POST", cfg.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create request", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.AgentToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AgentToken)
	}
	req.Header.Set("User-Agent", "Bitex-Agent/1.0")
	req.ContentLength = int64(contentLength)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Failed to send request", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Server returned error", map[string]string{"status": resp.Status})
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	logger.Info("Analysis sent successfully", map[string]string{"status": resp.Status})
	return nil
}
