// Package os provides operating system abstraction layer for Bitex
//
//nolint:revive // Package name 'os' is intentional for this internal abstraction layer
package os

import (
	stdos "os"
	"runtime"
	"time"

	"github.com/ilexum-group/bitex/internal/utils"
)

// OS defines the interface for operating system operations.
// This abstraction allows for different implementations per OS and easier testing.
type OS interface {
	// Hostname returns the host name reported by the kernel
	Hostname() (string, error)

	// Getenv retrieves the value of the environment variable named by the key
	Getenv(key string) string

	// Getwd returns a rooted path name corresponding to the current directory
	Getwd() (string, error)

	// Stat returns a FileInfo describing the named file
	Stat(name string) (stdos.FileInfo, error)

	// GetCurrentUser retrieves the current executing user based on the OS.
	GetCurrentUser() (string, error)

	// GetProcessID returns the current process ID.
	GetProcessID() int

	// SetLogger sets the logger function for OS operations.
	SetLogger(logger func(id, command string, args []string, startTime, endTime time.Time, exitCode int, err error, workingDirectory, targetResource string))
}

// DefaultImpl is the base implementation of Default with default methods.
type DefaultImpl struct {
	logger func(id, command string, args []string, startTime, endTime time.Time, exitCode int, err error, workingDirectory, targetResource string)
}

// NewDefault creates a new default OS implementation.
func NewDefault() *DefaultImpl {
	return &DefaultImpl{}
}

// Hostname returns the host name reported by the kernel.
func (d *DefaultImpl) Hostname() (string, error) {
	startTime := time.Now()
	hostname, err := stdos.Hostname()
	endTime := time.Now()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	if d.logger != nil {
		d.logger(utils.GenerateRandomID(), "os.Hostname", nil, startTime, endTime, exitCode, err, "", "")
	}
	return hostname, err
}

// Getenv retrieves the value of the environment variable named by the key.
func (d *DefaultImpl) Getenv(key string) string {
	startTime := time.Now()
	value := stdos.Getenv(key)
	endTime := time.Now()
	if d.logger != nil {
		d.logger(utils.GenerateRandomID(), "os.Getenv", []string{key}, startTime, endTime, 0, nil, "", "")
	}
	return value
}

// Getwd returns a rooted path name corresponding to the current directory.
func (d *DefaultImpl) Getwd() (string, error) {
	startTime := time.Now()
	path, err := stdos.Getwd()
	endTime := time.Now()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	if d.logger != nil {
		d.logger(utils.GenerateRandomID(), "os.Getwd", nil, startTime, endTime, exitCode, err, "", "")
	}
	return path, err
}

// Stat returns a FileInfo describing the named file.
func (d *DefaultImpl) Stat(name string) (stdos.FileInfo, error) {
	startTime := time.Now()
	fileInfo, err := stdos.Stat(name)
	endTime := time.Now()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	if d.logger != nil {
		d.logger(utils.GenerateRandomID(), "os.Stat", []string{name}, startTime, endTime, exitCode, err, "", name)
	}
	return fileInfo, err
}

// GetCurrentUser retrieves the current executing user (default implementation).
func (d *DefaultImpl) GetCurrentUser() (string, error) {
	return "unknown", nil
}

// GetProcessID returns the current process ID.
func (d *DefaultImpl) GetProcessID() int {
	startTime := time.Now()
	pid := stdos.Getpid()
	endTime := time.Now()
	if d.logger != nil {
		d.logger(utils.GenerateRandomID(), "os.GetProcessID", nil, startTime, endTime, 0, nil, "", "")
	}
	return pid
}

// SetLogger sets the logger function for OS operations.
func (d *DefaultImpl) SetLogger(logger func(id, command string, args []string, startTime, endTime time.Time, exitCode int, err error, workingDirectory, targetResource string)) {
	d.logger = logger
}

// New returns the appropriate OS implementation based on the runtime OS.
func New() OS {
	// Initialize with the appropriate OS-specific implementation
	switch runtime.GOOS {
	case "windows":
		return NewWindows()
	case "linux":
		return NewLinux()
	case "darwin":
		return NewDarwin()
	default:
		return NewDefault()
	}
}
