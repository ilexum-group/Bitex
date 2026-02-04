// Package utils provides shared utilities for Bitex.
//
//nolint:revive // utils is a common, acceptable package name for utility functions
package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateEvidenceID creates a unique evidence package identifier.
func GenerateEvidenceID(hostname string) string {
	randomID := GenerateRandomID()
	return fmt.Sprintf("BITX-%s-%s", hostname, randomID)
}

// GenerateRandomID creates a random identifier string UUID-like.
func GenerateRandomID() string {
	return uuid.New().String()
}
