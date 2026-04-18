package schools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	// DefaultPageLimit is applied when the caller does not specify page_limit
	// (or specifies 0). Matches the openapi.yaml default.
	DefaultPageLimit int32 = 25
	// MaxPageLimit caps the per-page result count. Matches the openapi.yaml maximum.
	MaxPageLimit int32 = 100
	// MaxSearchLength caps the q parameter length. Matches the openapi.yaml maxLength.
	// Typed int (not int32) because it's compared against len(string) at the
	// HTTP-boundary validator.
	MaxSearchLength int = 200
)

// ListSchoolsParams is the input to Service.ListSchools.
type ListSchoolsParams struct {
	Q      *string
	Limit  int32
	Cursor *Cursor
}

// ListSchoolsResult is the output of Service.ListSchools.
type ListSchoolsResult struct {
	Schools    []School
	HasMore    bool
	NextCursor *string
}

// Cursor is the opaque pagination token. Sorting is fixed on (name ASC, id ASC),
// so the cursor only needs the last consumed row's name and id.
type Cursor struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
}

// EncodeCursor serializes a Cursor into a base64-encoded string token.
func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("EncodeCursor: marshal: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses a base64-encoded string token back into a Cursor.
func DecodeCursor(s string) (Cursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: base64: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: json: %w", err)
	}
	return c, nil
}
