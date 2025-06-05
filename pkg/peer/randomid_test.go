package peer

import (
	"testing"
)

// TestRandomIDSuccess tests that randomID generates valid IDs
func TestRandomIDSuccess(t *testing.T) {
	id := randomID()
	if id == "" {
		t.Fatal("randomID returned empty string")
	}
	if len(id) != 32 { // 16 bytes hex encoded = 32 chars
		t.Fatalf("expected 32 character ID, got %d", len(id))
	}
	
	// Generate another ID and verify they're different
	id2 := randomID()
	if id == id2 {
		t.Fatal("randomID generated duplicate IDs")
	}
}