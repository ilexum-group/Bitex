// Package logger provides simple logging utilities for Bitex
package logger

import (
	"fmt"
	"os"
	"time"
)

// Info logs an informational message with optional fields.
func Info(msg string, fields map[string]string) {
	log("INFO", msg, fields)
}

// Error logs an error message with optional fields.
func Error(msg string, fields map[string]string) {
	log("ERROR", msg, fields)
}

// Debug logs a debug message with optional fields.
func Debug(msg string, fields map[string]string) {
	log("DEBUG", msg, fields)
}

func log(level, msg string, fields map[string]string) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(os.Stderr, "%s [%s] %s", timestamp, level, msg)
	if len(fields) > 0 {
		fmt.Fprintf(os.Stderr, " | ")
		for k, v := range fields {
			fmt.Fprintf(os.Stderr, "%s=%s ", k, v)
		}
	}
	fmt.Fprintln(os.Stderr)
}
