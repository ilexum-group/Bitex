// Package logger provides RFC 5424 compliant logging for Bitex
package logger

import (
	"fmt"
	"time"
)

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	// DebugLevel represents debug messages.
	DebugLevel LogLevel = 0
	// InfoLevel represents informational messages.
	InfoLevel LogLevel = 1
	// WarningLevel represents warning messages.
	WarningLevel LogLevel = 2
	// ErrorLevel represents error messages.
	ErrorLevel LogLevel = 3
)

// Logger defines the interface for logging operations.
type Logger interface {
	LogInfo(message string, meta map[string]string)
	LogWarn(message string, meta map[string]string)
	LogError(message string, meta map[string]string)
	LogDebug(message string, meta map[string]string)
}

// RFC5424Logger implements Logger with RFC 5424 compliant syslog format.
type RFC5424Logger struct {
	appName   string
	hostname  string
	processID string
}

// NewLogger creates a new RFC 5424 compliant logger.
func NewLogger(appName string, hostname string, processID string) (*RFC5424Logger, error) {
	if hostname == "" {
		hostname = "localhost"
	}

	return &RFC5424Logger{
		appName:   appName,
		hostname:  hostname,
		processID: processID,
	}, nil
}

// formatMessage creates an RFC 5424 formatted message.
func (l *RFC5424Logger) formatMessage(level LogLevel, message string, meta map[string]string) string {
	// RFC 5424 format: <PRI>VERSION TIMESTAMP HOSTNAME TAG PID MSGID STRUCTURED_DATA MESSAGE
	// We use facility 16 (local0) and severity level
	priority := (16 << 3) | int(level)
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	// Build structured data from metadata
	structuredData := "-"
	if len(meta) > 0 {
		structuredData = "[meta@1 "
		for key, value := range meta {
			// Escape special characters in value
			escapedValue := fmt.Sprintf("%q", value)
			structuredData += fmt.Sprintf("%s=%s ", key, escapedValue)
		}
		structuredData += "]"
	}

	// Format: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG
	logLine := fmt.Sprintf("<%d>1 %s %s %s %s - %s %s",
		priority, timestamp, l.hostname, l.appName, l.processID, structuredData, message)

	return logLine
}

// LogInfo logs an informational message.
func (l *RFC5424Logger) LogInfo(message string, meta map[string]string) {
	l.writeLog(InfoLevel, message, meta)
}

// LogWarn logs a warning message.
func (l *RFC5424Logger) LogWarn(message string, meta map[string]string) {
	l.writeLog(WarningLevel, message, meta)
}

// LogError logs an error message.
func (l *RFC5424Logger) LogError(message string, meta map[string]string) {
	l.writeLog(ErrorLevel, message, meta)
}

// LogDebug logs a debug message.
func (l *RFC5424Logger) LogDebug(message string, meta map[string]string) {
	l.writeLog(DebugLevel, message, meta)
}

// writeLog writes a log entry to stdout and captures it.
func (l *RFC5424Logger) writeLog(level LogLevel, message string, meta map[string]string) {
	formattedLog := l.formatMessage(level, message, meta)

	// Print to stdout
	fmt.Println(formattedLog)
}

// DefaultLogger is the global logger instance.
var DefaultLogger *RFC5424Logger

// InitDefaultLogger initializes the global logger.
func InitDefaultLogger(applicationName string, hostname string, processID string) error {
	logger, err := NewLogger(applicationName, hostname, processID)
	if err != nil {
		return err
	}
	DefaultLogger = logger
	return nil
}

// Convenience functions using the global logger

// LogInfo logs an informational message.
func LogInfo(message string, meta map[string]string) {
	if DefaultLogger != nil {
		DefaultLogger.LogInfo(message, meta)
	}
}

// LogWarn logs a warning message.
func LogWarn(message string, meta map[string]string) {
	if DefaultLogger != nil {
		DefaultLogger.LogWarn(message, meta)
	}
}

// LogError logs an error message.
func LogError(message string, meta map[string]string) {
	if DefaultLogger != nil {
		DefaultLogger.LogError(message, meta)
	}
}

// LogDebug logs a debug message.
func LogDebug(message string, meta map[string]string) {
	if DefaultLogger != nil {
		DefaultLogger.LogDebug(message, meta)
	}
}

// RFC5424 Severity levels.
const (
	Emergency = 0
	Alert     = 1
	Critical  = 2
	Error     = 3
	Warning   = 4
	Notice    = 5
	Info      = 6
	Debug     = 7
)
