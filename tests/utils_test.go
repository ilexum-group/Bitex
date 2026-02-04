package tests

import (
	"testing"

	"github.com/ilexum-group/bitex/internal/utils"
)

func TestGenerateRandomID(t *testing.T) {
	// Generate multiple IDs
	id1 := utils.GenerateRandomID()
	id2 := utils.GenerateRandomID()
	id3 := utils.GenerateRandomID()

	// Verify IDs are not empty
	if id1 == "" {
		t.Error("GenerateRandomID() should not return empty string")
	}
	if id2 == "" {
		t.Error("GenerateRandomID() should not return empty string")
	}
	if id3 == "" {
		t.Error("GenerateRandomID() should not return empty string")
	}

	// Verify IDs are unique
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
	if id2 == id3 {
		t.Error("Generated IDs should be unique")
	}
	if id1 == id3 {
		t.Error("Generated IDs should be unique")
	}

	// Verify ID format (UUID should be 36 characters with hyphens)
	if len(id1) != 36 {
		t.Errorf("ID length = %d, want 36 (UUID format)", len(id1))
	}
}

func TestGenerateRandomIDMultipleCalls(t *testing.T) {
	// Generate many IDs and verify uniqueness
	ids := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		id := utils.GenerateRandomID()

		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}

		ids[id] = true
	}

	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

func TestGenerateRandomIDFormat(t *testing.T) {
	id := utils.GenerateRandomID()

	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(id) != 36 {
		t.Fatalf("ID length = %d, want 36", len(id))
	}

	// Check hyphen positions
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		t.Errorf("ID format incorrect: %s", id)
	}

	// Verify all characters are valid hex or hyphens
	validChars := "0123456789abcdef-"
	for i, c := range id {
		found := false
		for _, valid := range validChars {
			if c == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Invalid character at position %d: %c in ID: %s", i, c, id)
		}
	}
}

func BenchmarkGenerateRandomID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = utils.GenerateRandomID()
	}
}

func TestGenerateRandomIDConcurrent(t *testing.T) {
	// Test concurrent ID generation
	iterations := 100
	ids := make(chan string, iterations)

	// Generate IDs concurrently
	for i := 0; i < iterations; i++ {
		go func() {
			ids <- utils.GenerateRandomID()
		}()
	}

	// Collect and verify uniqueness
	idMap := make(map[string]bool)
	for i := 0; i < iterations; i++ {
		id := <-ids
		if idMap[id] {
			t.Errorf("Duplicate ID in concurrent generation: %s", id)
		}
		idMap[id] = true
	}

	if len(idMap) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(idMap))
	}
}
