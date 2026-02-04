// Package sender handles sending analysis results to a remote server
package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ilexum-group/bitex/internal/logger"
	"github.com/ilexum-group/bitex/pkg/models"
)

// Sender encapsulates configuration for sending analysis to a remote server.
type Sender struct {
	serverURL  string
	authToken  string
	httpClient *http.Client
}

// NewSender builds a sender with the provided endpoint configuration.
func NewSender(serverURL, authToken string) *Sender {
	return &Sender{
		serverURL:  serverURL,
		authToken:  authToken,
		httpClient: &http.Client{},
	}
}

// WithHTTPClient overrides the HTTP client (useful for tests and custom transports).
func (s *Sender) WithHTTPClient(client *http.Client) *Sender {
	if client != nil {
		s.httpClient = client
	}
	return s
}

// SendAnalysis sends the TSK analysis result to a remote server.
func (s *Sender) SendAnalysis(analysis *models.TSKAnalysis) error {
	if analysis == nil {
		return fmt.Errorf("analysis is nil")
	}
	if s.serverURL == "" {
		return fmt.Errorf("server URL is required")
	}

	logger.LogInfo("Preparing to send analysis to server", map[string]string{
		"url":       s.serverURL,
		"disk_path": fmt.Sprintf("%d", len(analysis.DiskPath)),
	})

	// Strategy: Send as JSON
	jsonData, err := json.Marshal(analysis)
	if err != nil {
		logger.LogError("Failed to marshal analysis", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	contentLength := len(jsonData)
	logger.LogDebug("Sending analysis payload", map[string]string{
		"content_length": fmt.Sprintf("%d bytes", contentLength),
		"size_mb":        fmt.Sprintf("%.2f MB", float64(contentLength)/1024/1024),
	})

	return s.sendHTTPRequest(bytes.NewBuffer(jsonData), contentLength, "application/json")
}

// sendHTTPRequest performs the actual HTTP POST request.
func (s *Sender) sendHTTPRequest(body io.Reader, contentLength int, contentType string) error {
	if s.serverURL == "" {
		return fmt.Errorf("server URL is required")
	}

	req, err := http.NewRequest("POST", s.serverURL, body)
	if err != nil {
		logger.LogError("Failed to create HTTP request", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "Bitex-Agent/1.0")
	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}
	req.ContentLength = int64(contentLength)

	logger.LogDebug("Sending HTTP request", map[string]string{
		"method":            "POST",
		"content_type":      contentType,
		"content_length":    fmt.Sprintf("%d", contentLength),
		"content_length_mb": fmt.Sprintf("%.2f", float64(contentLength)/1024/1024),
	})

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("Failed to send HTTP request", map[string]string{"error": err.Error()})
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.LogError("Failed to close response body", map[string]string{"error": err.Error()})
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logger.LogWarn("Server returned non-OK status", map[string]string{
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		})
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	logger.LogInfo("Analysis successfully transmitted to server", map[string]string{
		"status_code": fmt.Sprintf("%d", resp.StatusCode),
	})

	return nil
}
