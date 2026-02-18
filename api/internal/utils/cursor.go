package utils

import (
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func CursorTimestamptz[C any](c *C, get func(*C) *time.Time) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{}
	}
	return Timestamptz(get(c))
}

func CursorText[C any](c *C, get func(*C) *string) pgtype.Text {
	if c == nil {
		return pgtype.Text{}
	}
	return Text(get(c))
}

func CursorInt8[C any](c *C, get func(*C) *int64) pgtype.Int8 {
	if c == nil {
		return pgtype.Int8{}
	}
	return Int8(get(c))
}

func CursorUUID[C any](c *C, get func(*C) [16]byte) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: get(c), Valid: true}
}

func CursorNullUploadStatus[C any](c *C, get func(*C) *string) db.NullUploadStatus {
	if c == nil {
		return db.NullUploadStatus{}
	}
	return NullUploadStatus(get(c))
}

func CursorNullMimeType[C any](c *C, get func(*C) *string) db.NullMimeType {
	if c == nil {
		return db.NullMimeType{}
	}
	return NullMimeType(get(c))
}