package tests

import (
	"runtime"
	"testing"
	"time"

	internalos "github.com/ilexum-group/bitex/internal/os"
)

func TestNew(t *testing.T) {
	osImpl := internalos.New()

	if osImpl == nil {
		t.Fatal("New() returned nil")
	}

	// Verify it returns the correct implementation for current OS
	switch runtime.GOOS {
	case "windows":
		if _, ok := osImpl.(*internalos.Windows); !ok {
			t.Errorf("Expected Windows for Windows, got %T", osImpl)
		}
	case "linux":
		if _, ok := osImpl.(*internalos.Linux); !ok {
			t.Errorf("Expected Linux for Linux, got %T", osImpl)
		}
	case "darwin":
		if _, ok := osImpl.(*internalos.Darwin); !ok {
			t.Errorf("Expected Darwin for Darwin, got %T", osImpl)
		}
	}
}

func TestHostname(t *testing.T) {
	osImpl := internalos.New()

	hostname, err := osImpl.Hostname()
	if err != nil {
		t.Errorf("Hostname() error = %v", err)
	}

	if hostname == "" {
		t.Error("Hostname should not be empty")
	}
}

func TestGetenv(t *testing.T) {
	osImpl := internalos.New()

	// Test with PATH which should exist on all systems
	path := osImpl.Getenv("PATH")

	// PATH should exist on all systems (might be empty but call should succeed)
	if path == "" {
		t.Log("PATH is empty (this might be expected in some test environments)")
	}
}

func TestGetwd(t *testing.T) {
	osImpl := internalos.New()

	wd, err := osImpl.Getwd()
	if err != nil {
		t.Errorf("Getwd() error = %v", err)
	}

	if wd == "" {
		t.Error("Working directory should not be empty")
	}
}

func TestStat(t *testing.T) {
	osImpl := internalos.New()

	// Test with current file
	fileInfo, err := osImpl.Stat("os_test.go")
	if err != nil {
		t.Errorf("Stat() error = %v", err)
	}

	if fileInfo == nil {
		t.Error("FileInfo should not be nil")
	}

	if fileInfo != nil && fileInfo.Name() != "os_test.go" {
		t.Errorf("Name() = %v, want os_test.go", fileInfo.Name())
	}

	// Test with non-existent file
	_, err = osImpl.Stat("nonexistent-file-12345.txt")
	if err == nil {
		t.Error("Stat() should return error for non-existent file")
	}
}

func TestGetCurrentUser(t *testing.T) {
	osImpl := internalos.New()

	user, err := osImpl.GetCurrentUser()
	if err != nil {
		t.Errorf("GetCurrentUser() error = %v", err)
	}

	if user == "" {
		t.Error("Current user should not be empty")
	}

	// On Windows, should get from USERNAME
	// On Linux/Darwin, should get from USER
	if runtime.GOOS == "windows" {
		t.Logf("Windows user: %s", user)
	} else {
		t.Logf("Unix user: %s", user)
	}
}

func TestGetProcessID(t *testing.T) {
	osImpl := internalos.New()

	pid := osImpl.GetProcessID()

	if pid <= 0 {
		t.Errorf("GetProcessID() = %v, should be positive", pid)
	}
}

func TestSetLogger(t *testing.T) {
	osImpl := internalos.New()

	logCalled := false
	testLogger := func(_, _ string, _ []string, _, _ time.Time, _ int, _ error, _, _ string) {
		logCalled = true
	}

	osImpl.SetLogger(testLogger)

	// Perform an operation that should trigger logging
	_, _ = osImpl.Hostname()

	if !logCalled {
		t.Error("Logger should have been called")
	}
}

func TestLoggerFunctionality(t *testing.T) {
	osImpl := internalos.New()

	var capturedCommand string
	var capturedArgs []string
	var capturedExitCode int

	testLogger := func(_, command string, args []string, _, _ time.Time, exitCode int, _ error, _, _ string) {
		capturedCommand = command
		capturedArgs = args
		capturedExitCode = exitCode
	}

	osImpl.SetLogger(testLogger)

	// Test Hostname logging
	_, _ = osImpl.Hostname()
	if capturedCommand != "os.Hostname" {
		t.Errorf("Command = %v, want os.Hostname", capturedCommand)
	}
	if capturedExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", capturedExitCode)
	}

	// Test Getenv logging
	_ = osImpl.Getenv("TEST_VAR")
	if capturedCommand != "os.Getenv" {
		t.Errorf("Command = %v, want os.Getenv", capturedCommand)
	}
	if len(capturedArgs) != 1 || capturedArgs[0] != "TEST_VAR" {
		t.Errorf("Args = %v, want [TEST_VAR]", capturedArgs)
	}

	// Test Stat logging with existing file
	_, _ = osImpl.Stat("os_test.go")
	if capturedCommand != "os.Stat" {
		t.Errorf("Command = %v, want os.Stat", capturedCommand)
	}

	// Test Stat logging with non-existent file (should have exitCode 1)
	_, _ = osImpl.Stat("nonexistent.txt")
	if capturedExitCode != 1 {
		t.Errorf("ExitCode for failed Stat = %v, want 1", capturedExitCode)
	}
}

func TestDefaultImplEmbedding(t *testing.T) {
	// Test that platform-specific implementations properly embed DefaultImpl
	osImpl := internalos.New()

	// All implementations should be able to call base methods
	_, err := osImpl.Hostname()
	if err != nil {
		t.Errorf("Embedded Hostname() error = %v", err)
	}

	_, err = osImpl.Getwd()
	if err != nil {
		t.Errorf("Embedded Getwd() error = %v", err)
	}

	_ = osImpl.Getenv("PATH")
}

func TestPlatformSpecificImplementations(t *testing.T) {
	// Test that each platform returns its specific implementation
	tests := []struct {
		goos     string
		wantType string
	}{
		{"windows", "*os.Windows"},
		{"linux", "*os.Linux"},
		{"darwin", "*os.Darwin"},
	}

	currentOS := runtime.GOOS

	for _, tt := range tests {
		if currentOS == tt.goos {
			osImpl := internalos.New()
			implType := ""

			switch osImpl.(type) {
			case *internalos.Windows:
				implType = "*os.Windows"
			case *internalos.Linux:
				implType = "*os.Linux"
			case *internalos.Darwin:
				implType = "*os.Darwin"
			}

			if implType != tt.wantType {
				t.Errorf("On %s, got %v, want %v", tt.goos, implType, tt.wantType)
			}
		}
	}
}

func TestNewDefault(t *testing.T) {
	impl := internalos.NewDefault()

	if impl == nil {
		t.Fatal("NewDefault() returned nil")
	}

	// Test that base methods work
	_, err := impl.Hostname()
	if err != nil {
		t.Errorf("Hostname() error = %v", err)
	}
}

func TestGetCurrentUserPlatformSpecific(t *testing.T) {
	osImpl := internalos.New()
	user, err := osImpl.GetCurrentUser()

	if err != nil {
		t.Errorf("GetCurrentUser() error = %v", err)
	}

	if user == "" {
		t.Error("User should not be empty")
	}

	// Verify it's not the default "unknown" value
	if user == "unknown" {
		t.Log("Warning: GetCurrentUser returned 'unknown' - environment variable may not be set")
	}
}

func TestLoggerWithNilLogger(t *testing.T) {
	osImpl := internalos.New()

	// Should not panic even with nil logger
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Operation panicked with nil logger: %v", r)
		}
	}()

	_, _ = osImpl.Hostname()
	_ = osImpl.Getenv("PATH")
	_, _ = osImpl.Getwd()
}

func TestStatWithValidPath(t *testing.T) {
	osImpl := internalos.New()

	// Test with current file
	fileInfo, err := osImpl.Stat("os_test.go")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if fileInfo.IsDir() {
		t.Error("os_test.go should not be a directory")
	}

	if fileInfo.Size() <= 0 {
		t.Error("File size should be greater than 0")
	}
}

func TestMultipleOperationsWithLogging(t *testing.T) {
	osImpl := internalos.New()

	callCount := 0
	testLogger := func(_, _ string, _ []string, startTime, endTime time.Time, _ int, _ error, _, _ string) {
		callCount++

		// Verify timing
		if endTime.Before(startTime) {
			t.Error("EndTime should be after StartTime")
		}
	}

	osImpl.SetLogger(testLogger)

	// Perform multiple operations
	_, _ = osImpl.Hostname()
	_ = osImpl.Getenv("PATH")
	_, _ = osImpl.Getwd()
	_, _ = osImpl.Stat("os_test.go")

	expectedCalls := 4
	if callCount != expectedCalls {
		t.Errorf("Logger called %d times, want %d", callCount, expectedCalls)
	}
}
