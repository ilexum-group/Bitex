package tests

import (
	"testing"

	"github.com/ilexum-group/bitex/internal/logger"
)

func TestInitDefaultLogger(t *testing.T) {
	tests := []struct {
		name      string
		appName   string
		hostname  string
		processID string
	}{
		{
			name:      "Valid initialization",
			appName:   "TestApp",
			hostname:  "test-host",
			processID: "12345",
		},
		{
			name:      "Empty app name",
			appName:   "",
			hostname:  "test-host",
			processID: "12345",
		},
		{
			name:      "Empty hostname",
			appName:   "TestApp",
			hostname:  "",
			processID: "12345",
		},
		{
			name:      "Empty process ID",
			appName:   "TestApp",
			hostname:  "test-host",
			processID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("InitDefaultLogger() panicked: %v", r)
				}
			}()

			if err := logger.InitDefaultLogger(tt.appName, tt.hostname, tt.processID); err != nil {
				t.Errorf("InitDefaultLogger() error = %v", err)
			}
		})
	}
}

func TestLogInfo(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	tests := []struct {
		name    string
		message string
		context map[string]string
	}{
		{
			name:    "Simple log",
			message: "Test message",
			context: nil,
		},
		{
			name:    "Log with context",
			message: "Test with context",
			context: map[string]string{"key": "value", "count": "42"},
		},
		{
			name:    "Log with empty context",
			message: "Empty context",
			context: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogInfo() panicked: %v", r)
				}
			}()

			logger.LogInfo(tt.message, tt.context)
		})
	}
}

func TestLogWarning(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	tests := []struct {
		name    string
		message string
		context map[string]string
	}{
		{
			name:    "Simple warning",
			message: "Warning message",
			context: nil,
		},
		{
			name:    "Warning with context",
			message: "Warning with details",
			context: map[string]string{"reason": "test", "severity": "low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogWarning() panicked: %v", r)
				}
			}()

			logger.LogWarn(tt.message, tt.context)
		})
	}
}

func TestLogError(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	tests := []struct {
		name    string
		message string
		context map[string]string
	}{
		{
			name:    "Simple error",
			message: "Error occurred",
			context: nil,
		},
		{
			name:    "Error with context",
			message: "Detailed error",
			context: map[string]string{"error": "file not found", "path": "/tmp/test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogError() panicked: %v", r)
				}
			}()

			logger.LogError(tt.message, tt.context)
		})
	}
}

func TestLogDebug(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	tests := []struct {
		name    string
		message string
		context map[string]string
	}{
		{
			name:    "Simple debug",
			message: "Debug message",
			context: nil,
		},
		{
			name:    "Debug with context",
			message: "Debug details",
			context: map[string]string{"variable": "value", "state": "running"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogDebug() panicked: %v", r)
				}
			}()

			logger.LogDebug(tt.message, tt.context)
		})
	}
}

func TestMultipleLogCalls(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	// Should handle multiple consecutive log calls without issues
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Multiple log calls panicked: %v", r)
		}
	}()

	logger.LogInfo("First message", nil)
	logger.LogWarn("Second message", nil)
	logger.LogError("Third message", nil)
	logger.LogDebug("Fourth message", nil)
	logger.LogInfo("Fifth message", map[string]string{"count": "5"})
}

func TestLogWithLargeContext(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	// Create large context
	context := make(map[string]string)
	for i := 0; i < 100; i++ {
		context[string(rune(i))] = "value"
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log with large context panicked: %v", r)
		}
	}()

	logger.LogInfo("Large context", context)
}

func TestLogWithSpecialCharacters(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	tests := []struct {
		name    string
		message string
		context map[string]string
	}{
		{
			name:    "Newlines in message",
			message: "Line 1\nLine 2\nLine 3",
			context: nil,
		},
		{
			name:    "Special characters",
			message: "Test: @#$%^&*()[]{}|\\",
			context: map[string]string{"special": "!@#$%"},
		},
		{
			name:    "Unicode characters",
			message: "Unicode: 日本語 español 中文",
			context: map[string]string{"emoji": "🔍🔒"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Log with special characters panicked: %v", r)
				}
			}()

			logger.LogInfo(tt.message, tt.context)
		})
	}
}

func BenchmarkLogInfo(b *testing.B) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		b.Fatalf("InitDefaultLogger() error = %v", err)
	}
	context := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogInfo("Benchmark message", context)
	}
}

func BenchmarkLogError(b *testing.B) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		b.Fatalf("InitDefaultLogger() error = %v", err)
	}
	context := map[string]string{"error": "test error"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogError("Benchmark error", context)
	}
}

func TestConcurrentLogging(t *testing.T) {
	if err := logger.InitDefaultLogger("TestApp", "test-host", "12345"); err != nil {
		t.Fatalf("InitDefaultLogger() error = %v", err)
	}

	// Test concurrent logging
	done := make(chan bool)
	goroutines := 10
	iterations := 100

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < iterations; j++ {
				logger.LogInfo("Concurrent log", map[string]string{"goroutine": string(rune(id))})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
