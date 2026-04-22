package recents

import "github.com/google/uuid"

const (
	// DefaultLimit is applied when the caller does not specify a
	// limit query parameter (or specifies 0). Matches the openapi.yaml
	// default.
	DefaultLimit int32 = 10
	// MinLimit is the inclusive lower bound on limit. Matches openapi.yaml.
	MinLimit int32 = 1
	// MaxLimit is the inclusive upper bound on limit. Matches openapi.yaml.
	MaxLimit int32 = 30
)

// ListRecentsParams is the input to Service.ListRecents. ViewerID is
// resolved from the Clerk JWT in the handler before this struct is
// constructed; the service never sees the JWT directly.
type ListRecentsParams struct {
	ViewerID uuid.UUID
	Limit    int32
}

// ListRecentsResult is the output of Service.ListRecents. A struct
// rather than a bare slice so future additions (an echoed limit, a
// computed total_view_count) can land backwards-compatibly.
type ListRecentsResult struct {
	Recents []RecentItem
}
