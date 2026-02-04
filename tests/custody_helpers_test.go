package tests

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ilexum-group/bitex/pkg/models"
)

func TestNewCustodyChainEntry(t *testing.T) {
	tests := []struct {
		name      string
		agentType string
		version   string
		wantErr   bool
	}{
		{
			name:      "Valid entry - Bitex",
			agentType: "bitex",
			version:   "1.0.0",
			wantErr:   false,
		},
		{
			name:      "Valid entry - Evidex",
			agentType: "evidex",
			version:   "2.1.0",
			wantErr:   false,
		},
		{
			name:      "Valid entry - Empty values",
			agentType: "",
			version:   "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := models.NewCustodyChainEntry(tt.agentType, tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustodyChainEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if entry.ID == "" {
					t.Error("NewCustodyChainEntry() ID should not be empty")
				}
				if entry.AgentType != tt.agentType {
					t.Errorf("AgentType = %v, want %v", entry.AgentType, tt.agentType)
				}
				if entry.AgentVersion != tt.version {
					t.Errorf("AgentVersion = %v, want %v", entry.AgentVersion, tt.version)
				}
				if entry.LogEntries == nil {
					t.Error("LogEntries should be initialized")
				}
				if entry.CommandHistory == nil {
					t.Error("CommandHistory should be initialized")
				}
				if entry.StartTimestamp.IsZero() {
					t.Error("StartTimestamp should be set")
				}
			}
		})
	}
}

func TestLogEntries(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	// Test LogInfo
	initialCount := len(entry.LogEntries)
	entry.LogInfo("test_operation", "Test info message")

	if len(entry.LogEntries) != initialCount+1 {
		t.Errorf("Expected %d log entries, got %d", initialCount+1, len(entry.LogEntries))
	}

	lastEntry := entry.LogEntries[len(entry.LogEntries)-1]
	if lastEntry.Level != models.LogLevelInfo {
		t.Errorf("Level = %v, want %v", lastEntry.Level, models.LogLevelInfo)
	}
	if lastEntry.Message != "Test info message" {
		t.Errorf("Message = %v, want %v", lastEntry.Message, "Test info message")
	}
	if lastEntry.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestLogCommandExecution(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	startTime := time.Now().Add(-5 * time.Second)
	endTime := time.Now()

	entry.LogCommand("cmd-123", "mmls", []string{"/dev/sda"}, startTime, endTime, 0, nil, "/tmp", "/dev/sda")

	if len(entry.CommandHistory) != 1 {
		t.Errorf("Expected 1 command in history, got %d", len(entry.CommandHistory))
	}

	lastCmd := entry.CommandHistory[0]
	if lastCmd.ID != "cmd-123" {
		t.Errorf("ID = %v, want %v", lastCmd.ID, "cmd-123")
	}
	if lastCmd.Command != "mmls" {
		t.Errorf("Command = %v, want %v", lastCmd.Command, "mmls")
	}
}

func TestLogCommand(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	startTime := time.Now().Add(-2 * time.Second)
	endTime := time.Now()

	entry.LogCommand("cmd-456", "fsstat", []string{"-f", "ext4", "/dev/sda1"},
		startTime, endTime, 0, nil, "/tmp", "/dev/sda1")

	if len(entry.CommandHistory) != 1 {
		t.Errorf("Expected 1 command in history, got %d", len(entry.CommandHistory))
	}

	cmd := entry.CommandHistory[0]
	if cmd.ID != "cmd-456" {
		t.Errorf("ID = %v, want %v", cmd.ID, "cmd-456")
	}
	if cmd.Command != "fsstat" {
		t.Errorf("Command = %v, want %v", cmd.Command, "fsstat")
	}
	if cmd.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want %v", cmd.ExitCode, 0)
	}
	if cmd.WorkingDirectory != "/tmp" {
		t.Errorf("WorkingDirectory = %v, want %v", cmd.WorkingDirectory, "/tmp")
	}
}

func TestFinalizeFromReader(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	// Create test data
	testData := "This is test data for hashing"
	reader := bytes.NewBufferString(testData)

	err := entry.FinalizeFromReader(reader, 10)
	if err != nil {
		t.Errorf("FinalizeFromReader() error = %v", err)
	}

	// Verify timestamps
	if entry.EndTimestamp.IsZero() {
		t.Error("EndTimestamp should be set")
	}
	if entry.Duration == "" {
		t.Error("Duration should not be empty")
	}

	// Verify size
	expectedSize := int64(len(testData))
	if entry.TotalSizeBytes != expectedSize {
		t.Errorf("TotalSizeBytes = %v, want %v", entry.TotalSizeBytes, expectedSize)
	}

	// Verify item count
	if entry.ItemCount != 10 {
		t.Errorf("ItemCount = %v, want %v", entry.ItemCount, 10)
	}

	// Verify hashes are not empty
	if entry.MD5Hash == "" {
		t.Error("MD5Hash should not be empty")
	}
	if entry.SHA1Hash == "" {
		t.Error("SHA1Hash should not be empty")
	}
	if entry.SHA256Hash == "" {
		t.Error("SHA256Hash should not be empty")
	}

	// Verify hash format (hex string)
	if len(entry.MD5Hash) != 32 {
		t.Errorf("MD5Hash length = %v, want 32", len(entry.MD5Hash))
	}
	if len(entry.SHA1Hash) != 40 {
		t.Errorf("SHA1Hash length = %v, want 40", len(entry.SHA1Hash))
	}
	if len(entry.SHA256Hash) != 64 {
		t.Errorf("SHA256Hash length = %v, want 64", len(entry.SHA256Hash))
	}
}

func TestFinalizeFromReaderEmptyData(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	// Create empty reader
	reader := bytes.NewBufferString("")

	err := entry.FinalizeFromReader(reader, 0)
	if err != nil {
		t.Errorf("FinalizeFromReader() error = %v", err)
	}

	if entry.TotalSizeBytes != 0 {
		t.Errorf("TotalSizeBytes = %v, want 0", entry.TotalSizeBytes)
	}

	// Even empty data should produce valid hashes
	if entry.MD5Hash == "" {
		t.Error("MD5Hash should not be empty even for empty data")
	}
}

func TestMultipleLogEntries(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	// Add multiple log entries
	entry.LogInfo("op1", "First message")
	entry.LogWarning("op2", "Second message")
	entry.LogError("op3", "Third message", nil)

	if len(entry.LogEntries) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(entry.LogEntries))
	}

	// Verify order
	if entry.LogEntries[0].Message != "First message" {
		t.Error("First entry should be 'First message'")
	}
	if entry.LogEntries[1].Message != "Second message" {
		t.Error("Second entry should be 'Second message'")
	}
	if entry.LogEntries[2].Message != "Third message" {
		t.Error("Third entry should be 'Third message'")
	}
}

func TestMultipleCommands(t *testing.T) {
	entry, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	// Add multiple commands
	entry.LogCommand("cmd-1", "mmls", []string{"/dev/sda"}, time.Now(), time.Now(), 0, nil, "/tmp", "/dev/sda")
	entry.LogCommand("cmd-2", "fsstat", []string{"/dev/sda1"}, time.Now(), time.Now(), 0, nil, "/tmp", "/dev/sda1")
	entry.LogCommand("cmd-3", "fls", []string{"-r", "/dev/sda1"}, time.Now(), time.Now(), 0, nil, "/tmp", "/dev/sda1")

	if len(entry.CommandHistory) != 3 {
		t.Errorf("Expected 3 commands in history, got %d", len(entry.CommandHistory))
	}

	// Verify commands are in order
	if entry.CommandHistory[0].Command != "mmls" {
		t.Error("First command should be 'mmls'")
	}
	if entry.CommandHistory[1].Command != "fsstat" {
		t.Error("Second command should be 'fsstat'")
	}
	if entry.CommandHistory[2].Command != "fls" {
		t.Error("Third command should be 'fls'")
	}
}

func TestHashConsistency(t *testing.T) {
	// Create two entries with same data
	entry1, _ := models.NewCustodyChainEntry("bitex", "1.0.0")
	entry2, _ := models.NewCustodyChainEntry("bitex", "1.0.0")

	testData := "Consistent test data"
	reader1 := strings.NewReader(testData)
	reader2 := strings.NewReader(testData)

	if err := entry1.FinalizeFromReader(reader1, 1); err != nil {
		t.Fatalf("entry1.FinalizeFromReader() error = %v", err)
	}
	if err := entry2.FinalizeFromReader(reader2, 1); err != nil {
		t.Fatalf("entry2.FinalizeFromReader() error = %v", err)
	}

	// Hashes should be identical for same data
	if entry1.MD5Hash != entry2.MD5Hash {
		t.Error("MD5 hashes should match for identical data")
	}
	if entry1.SHA1Hash != entry2.SHA1Hash {
		t.Error("SHA1 hashes should match for identical data")
	}
	if entry1.SHA256Hash != entry2.SHA256Hash {
		t.Error("SHA256 hashes should match for identical data")
	}
}
