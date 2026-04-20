package recents

import (
	"context"
	"fmt"
	"sort"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

// Service implements the GET /api/me/recents business logic.
type Service struct {
	repo Repository
}

// NewService wires a recents Service over the given repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListRecents returns the viewer's most recently viewed entities,
// merged across files, study guides, and courses (ASK-145).
//
// Strategy:
//   - Fan out to three independent sqlc queries (one per entity
//     type), each capped at the requested limit. A SQL UNION would
//     force a lowest-common-denominator projection or per-row
//     casting; in-process merge keeps each query hitting its own
//     focused (user_id, viewed_at) index and lets the mapper produce
//     a discriminated-union shape.
//   - Sort the combined slice by ViewedAt DESC, breaking ties on
//     EntityID lexicographic so the output is deterministic when
//     two views landed in the same microsecond.
//   - Truncate to limit. Each per-type query already returned at most
//     `limit` rows, so the unsorted combined slice has at most
//     3*limit entries -- the merge is O(N log N) in N=3*limit, which
//     is a negligible amount of work at MaxLimit=30 (90 items).
//
// Limit is applied per-query AND post-merge because either bound
// alone is wrong: a per-query bound alone returns up to 3*limit
// items when the caller asked for limit; a post-merge bound alone
// would require fetching every row from each table to guarantee the
// limit-most-recent global window. Both bounds together are correct
// and bounded.
func (s *Service) ListRecents(ctx context.Context, p ListRecentsParams) (ListRecentsResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit < MinLimit || limit > MaxLimit {
		return ListRecentsResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"limit": fmt.Sprintf("must be between %d and %d", MinLimit, MaxLimit),
		})
	}

	viewerPgxID := utils.UUID(p.ViewerID)

	// Three independent reads; if any one fails the whole request
	// fails (per the spec's "all three must succeed" Dependency
	// Failures contract). No ergonomic gain from goroutine fan-out at
	// this scale: the queries are <2ms each on indexed reads, the
	// connection-pool churn outweighs the parallelism win.
	fileRows, err := s.repo.ListRecentFiles(ctx, db.ListRecentFilesParams{
		ViewerID:  viewerPgxID,
		PageLimit: limit,
	})
	if err != nil {
		return ListRecentsResult{}, fmt.Errorf("ListRecents: files: %w", err)
	}
	guideRows, err := s.repo.ListRecentStudyGuides(ctx, db.ListRecentStudyGuidesParams{
		ViewerID:  viewerPgxID,
		PageLimit: limit,
	})
	if err != nil {
		return ListRecentsResult{}, fmt.Errorf("ListRecents: study guides: %w", err)
	}
	courseRows, err := s.repo.ListRecentCourses(ctx, db.ListRecentCoursesParams{
		ViewerID:  viewerPgxID,
		PageLimit: limit,
	})
	if err != nil {
		return ListRecentsResult{}, fmt.Errorf("ListRecents: courses: %w", err)
	}

	// Pre-size to the sum to avoid append regrowth.
	merged := make([]RecentItem, 0, len(fileRows)+len(guideRows)+len(courseRows))
	for _, r := range fileRows {
		item, err := mapRecentFile(r)
		if err != nil {
			return ListRecentsResult{}, fmt.Errorf("ListRecents: %w", err)
		}
		merged = append(merged, item)
	}
	for _, r := range guideRows {
		item, err := mapRecentStudyGuide(r)
		if err != nil {
			return ListRecentsResult{}, fmt.Errorf("ListRecents: %w", err)
		}
		merged = append(merged, item)
	}
	for _, r := range courseRows {
		item, err := mapRecentCourse(r)
		if err != nil {
			return ListRecentsResult{}, fmt.Errorf("ListRecents: %w", err)
		}
		merged = append(merged, item)
	}

	// ViewedAt DESC, then EntityID lexicographic to make ties
	// deterministic. EntityID.String() formats the canonical
	// 8-4-4-4-12 lower-case hex form, which is byte-stable for the
	// purposes of comparison.
	sort.SliceStable(merged, func(i, j int) bool {
		if !merged[i].ViewedAt.Equal(merged[j].ViewedAt) {
			return merged[i].ViewedAt.After(merged[j].ViewedAt)
		}
		return merged[i].EntityID.String() < merged[j].EntityID.String()
	})

	if int32(len(merged)) > limit {
		merged = merged[:limit]
	}

	return ListRecentsResult{Recents: merged}, nil
}
