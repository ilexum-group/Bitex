// Package os provides operating system abstraction layer for Bitex
//
//nolint:revive // Package name 'os' is intentional for this internal abstraction layer
package os

// Darwin is the macOS (Darwin)-specific implementation.
type Darwin struct {
	*DefaultImpl // Embed default implementation
}

// NewDarwin creates a new macOS-specific OS implementation.
func NewDarwin() OS {
	return &Darwin{
		DefaultImpl: NewDefault(),
	}
}

// GetCurrentUser retrieves the current executing user on macOS.
func (d *Darwin) GetCurrentUser() (string, error) {
	username := d.Getenv("USER")
	if username == "" {
		return "unknown", nil
	}
	return username, nil
}
