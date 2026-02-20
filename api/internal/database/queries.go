package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUID wraps google uuid for backward compatibility.
type UUID = uuid.UUID

// PgUUID converts a google uuid to pgtype.UUID.
func PgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// UUIDStr converts a pgtype.UUID to string.
func UUIDStr(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}

// TextPtr converts pgtype.Text to *string.
func TextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// Int4Ptr converts pgtype.Int4 to *int32.
func Int4Ptr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// PgText converts a string to pgtype.Text.
func PgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// PgTextPtr converts a *string to pgtype.Text.
func PgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// PgTimePtr converts a *time.Time to pgtype.Timestamptz.
func PgTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
