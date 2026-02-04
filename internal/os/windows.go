// Package os provides operating system abstraction layer for Bitex
//
//nolint:revive // Package name 'os' is intentional for this internal abstraction layer
package os

// Windows is the Windows-specific implementation.
type Windows struct {
	*DefaultImpl // Embed default implementation
}

// NewWindows creates a new Windows-specific OS implementation.
func NewWindows() OS {
	return &Windows{
		DefaultImpl: NewDefault(),
	}
}

// GetCurrentUser retrieves the current executing user on Windows.
func (w *Windows) GetCurrentUser() (string, error) {
	username := w.Getenv("USERNAME")
	if username == "" {
		return "unknown", nil
	}
	return username, nil
}
