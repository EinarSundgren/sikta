package extraction

import (
	"github.com/google/uuid"
)

// parseUUID converts a string to uuid.UUID.
func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// stringPtr returns a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
