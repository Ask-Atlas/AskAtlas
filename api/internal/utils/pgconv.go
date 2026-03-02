package utils

import (
	"fmt"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Text converts a string pointer to a pgtype.Text.
func Text(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// Timestamptz converts a time.Time pointer to a pgtype.Timestamptz.
func Timestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// Int8 converts an int64 pointer to a pgtype.Int8.
func Int8(n *int64) pgtype.Int8 {
	if n == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *n, Valid: true}
}

// UUID converts a byte array to a pgtype.UUID.
func UUID(b [16]byte) pgtype.UUID {
	return pgtype.UUID{Bytes: b, Valid: true}
}

// NullUploadStatus converts a string pointer to a db.NullUploadStatus.
func NullUploadStatus(s *string) db.NullUploadStatus {
	if s == nil {
		return db.NullUploadStatus{}
	}
	return db.NullUploadStatus{UploadStatus: db.UploadStatus(*s), Valid: true}
}

// NullMimeType converts a string pointer to a db.NullMimeType.
func NullMimeType(s *string) db.NullMimeType {
	if s == nil {
		return db.NullMimeType{}
	}
	return db.NullMimeType{MimeType: db.MimeType(*s), Valid: true}
}

// TimestamptzPtr converts a pgtype.Timestamptz to a time.Time pointer.
func TimestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// PgxToGoogleUUID converts a pgtype.UUID to a standard Google UUID.
func PgxToGoogleUUID(u pgtype.UUID) (uuid.UUID, error) {
	if !u.Valid {
		return uuid.Nil, fmt.Errorf("uuid is NULL/invalid")
	}
	return uuid.FromBytes(u.Bytes[:])
}
