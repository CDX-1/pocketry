package auth

import (
	"encoding/base64"
	"testing"
)

// Tests that NewFlowID generates a non-empty flow ID.
func TestNewFlowID(t *testing.T) {
	id, err := NewFlowID()
	if err != nil {
		t.Fatalf("NewFlowID returned error: %v", err)
	}

	if id == "" {
		t.Fatal("expected non-empty flow ID")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		t.Fatalf("flow ID is not valid raw base64url: %v", err)
	}

	if len(decoded) != 32 {
		t.Fatalf(
			"expected flow ID to decode to 32 bytes, got %d",
			len(decoded),
		)
	}
}

// Tests that NewFlowID produces unique IDs.
func TestNewFlowIDProducesUniqueIDs(t *testing.T) {
	first, err := NewFlowID()
	if err != nil {
		t.Fatalf("first NewFlowID returned error: %v", err)
	}

	second, err := NewFlowID()
	if err != nil {
		t.Fatalf("second NewFlowID returned error: %v", err)
	}

	if first == second {
		t.Fatal("expected generated flow IDs to differ")
	}
}

// Tests that NewFlowID uses raw base64url without padding.
func TestNewFlowIDUsesRawBase64URLWithoutPadding(t *testing.T) {
	id, err := NewFlowID()
	if err != nil {
		t.Fatalf("NewFlowID returned error: %v", err)
	}

	for _, character := range id {
		switch {
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case character >= '0' && character <= '9':
		case character == '-':
		case character == '_':
		default:
			t.Fatalf(
				"flow ID contains invalid base64url character %q",
				character,
			)
		}
	}

	if len(id) != 43 {
		t.Fatalf(
			"expected 43-character raw base64url flow ID, got %d",
			len(id),
		)
	}
}