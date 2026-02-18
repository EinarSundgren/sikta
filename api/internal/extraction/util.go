package extraction

import (
	"github.com/google/uuid"
	"github.com/einarsundgren/sikta/internal/database"
)

// parseUUID converts a string to database.UUID.
func parseUUID(s string) database.UUID {
	id, _ := uuid.Parse(s)
	return database.UUID(id)
}

// stringPtr returns a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
