package favorites

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

const (
	// DefaultLimit is applied when the caller omits the limit query
	// param (or specifies 0). Matches the openapi.yaml default.
	DefaultLimit int32 = 25
	// MinLimit is the inclusive lower bound on limit. Matches openapi.yaml.
	MinLimit int32 = 1
	// MaxLimit is the inclusive upper bound on limit. Matches openapi.yaml.
	MaxLimit int32 = 100

	// MaxOffset caps the per-table read width. With MaxLimit=100 and
	// MaxOffset=1000 the worst case is 1101 rows per table * 3 tables
	// = ~3.3k rows in memory -- plenty of headroom for users with
	// many favorites while still bounding the per-request cost.
	// A request with offset > MaxOffset returns 400 with
	// "invalid cursor value".
	MaxOffset int32 = 1000
)

// ListFavoritesParams is the input to Service.ListFavorites.
// EntityType filters to a single type when non-nil; nil means all
// three types are merged.
type ListFavoritesParams struct {
	ViewerID   uuid.UUID
	EntityType *EntityType
	Limit      int32
	Cursor     *string
}

// ListFavoritesResult is the output of Service.ListFavorites.
// NextCursor is *string so an explicit JSON null on the last page
// can be distinguished from an absent field on the wire.
type ListFavoritesResult struct {
	Favorites  []FavoriteItem
	HasMore    bool
	NextCursor *string
}

// EncodeCursor wraps a non-negative integer offset in a base64
// envelope. The wire contract is opaque -- a future migration to
// keyset pagination is non-breaking because clients pass cursors
// back verbatim and don't parse them.
func EncodeCursor(offset int32) string {
	return base64.URLEncoding.EncodeToString([]byte(strconv.FormatInt(int64(offset), 10)))
}

// DecodeCursor parses an opaque cursor string into an integer offset.
// Returns an error for malformed base64, non-integer payload, or a
// negative / over-cap offset; the handler maps that to 400 with the
// spec's "invalid cursor value" detail.
func DecodeCursor(s string) (int32, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return 0, fmt.Errorf("DecodeCursor: base64: %w", err)
	}
	n, err := strconv.ParseInt(string(raw), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("DecodeCursor: parse int: %w", err)
	}
	if n < 0 {
		return 0, fmt.Errorf("DecodeCursor: negative offset %d", n)
	}
	if int32(n) > MaxOffset {
		return 0, fmt.Errorf("DecodeCursor: offset %d exceeds MaxOffset %d", n, MaxOffset)
	}
	return int32(n), nil
}
