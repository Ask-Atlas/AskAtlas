// Package utils contains shared helper functions used throughout the application.
package utils

import (
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// CursorTimestamptz creates a pgtype.Timestamptz extracted from a paginated cursor.
func CursorTimestamptz[C any](c *C, get func(*C) *time.Time) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{}
	}
	return Timestamptz(get(c))
}

// CursorText creates a pgtype.Text extracted from a paginated cursor.
func CursorText[C any](c *C, get func(*C) *string) pgtype.Text {
	if c == nil {
		return pgtype.Text{}
	}
	return Text(get(c))
}

// CursorInt8 creates a pgtype.Int8 extracted from a paginated cursor.
func CursorInt8[C any](c *C, get func(*C) *int64) pgtype.Int8 {
	if c == nil {
		return pgtype.Int8{}
	}
	return Int8(get(c))
}

// CursorUUID creates a pgtype.UUID extracted from a paginated cursor.
func CursorUUID[C any](c *C, get func(*C) [16]byte) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: get(c), Valid: true}
}

// CursorNullUploadStatus creates a db.NullUploadStatus extracted from a paginated cursor.
func CursorNullUploadStatus[C any](c *C, get func(*C) *string) db.NullUploadStatus {
	if c == nil {
		return db.NullUploadStatus{}
	}
	return NullUploadStatus(get(c))
}

