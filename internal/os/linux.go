// Package os provides operating system abstraction layer for Bitex
//
//nolint:revive // Package name 'os' is intentional for this internal abstraction layer
package os

// Linux is the Linux-specific implementation.
type Linux struct {
	*DefaultImpl // Embed default implementation
}

// NewLinux creates a new Linux-specific OS implementation.
func NewLinux() OS {
	return &Linux{
		DefaultImpl: NewDefault(),
	}
}

// GetCurrentUser retrieves the current executing user on Linux.
func (l *Linux) GetCurrentUser() (string, error) {
	username := l.Getenv("USER")
	if username == "" {
		return "unknown", nil
	}
	return username, nil
}
