package favorites

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToggleFavoriteResult is the domain output of the per-entity
// favorite-toggle endpoints (ASK-130 / ASK-156 / ASK-157).
// `Favorited` is true when the toggle inserted a row, false when it
// deleted one. `FavoritedAt` carries the row timestamp on insert and
// is nil on delete (renders as JSON null on the wire).
type ToggleFavoriteResult struct {
	Favorited   bool
	FavoritedAt *time.Time
}

// ToggleFileFavorite toggles the (viewer, file) row in
// file_favorites (ASK-130). Returns apperrors.ErrNotFound when the
// file is missing or in any deletion state -- favoriting is
// permission-less but requires a live parent row.
func (s *Service) ToggleFileFavorite(ctx context.Context, viewerID, fileID uuid.UUID) (ToggleFavoriteResult, error) {
	if err := s.repo.CheckFileExists(ctx, utils.UUID(fileID)); err != nil {
		return ToggleFavoriteResult{}, err
	}
	row, err := s.repo.ToggleFileFavorite(ctx, db.ToggleFileFavoriteParams{
		UserID: utils.UUID(viewerID),
		FileID: utils.UUID(fileID),
	})
	if err != nil {
		return ToggleFavoriteResult{}, fmt.Errorf("ToggleFileFavorite: %w", err)
	}
	return ToggleFavoriteResult{
		Favorited:   row.Favorited,
		FavoritedAt: utils.TimestamptzPtr(row.FavoritedAt),
	}, nil
}

// ToggleStudyGuideFavorite toggles the (viewer, study_guide) row in
// study_guide_favorites (ASK-156). Same 404 rules as the file path:
// missing or soft-deleted (deleted_at IS NOT NULL) maps to ErrNotFound.
func (s *Service) ToggleStudyGuideFavorite(ctx context.Context, viewerID, studyGuideID uuid.UUID) (ToggleFavoriteResult, error) {
	if err := s.repo.CheckStudyGuideExists(ctx, utils.UUID(studyGuideID)); err != nil {
		return ToggleFavoriteResult{}, err
	}
	row, err := s.repo.ToggleStudyGuideFavorite(ctx, db.ToggleStudyGuideFavoriteParams{
		UserID:       utils.UUID(viewerID),
		StudyGuideID: utils.UUID(studyGuideID),
	})
	if err != nil {
		return ToggleFavoriteResult{}, fmt.Errorf("ToggleStudyGuideFavorite: %w", err)
	}
	return ToggleFavoriteResult{
		Favorited:   row.Favorited,
		FavoritedAt: utils.TimestamptzPtr(row.FavoritedAt),
	}, nil
}

// ToggleCourseFavorite toggles the (viewer, course) row in
// course_favorites (ASK-157). Courses do not support soft-delete so
// existence is the only 404 condition.
func (s *Service) ToggleCourseFavorite(ctx context.Context, viewerID, courseID uuid.UUID) (ToggleFavoriteResult, error) {
	if err := s.repo.CheckCourseExists(ctx, utils.UUID(courseID)); err != nil {
		return ToggleFavoriteResult{}, err
	}
	row, err := s.repo.ToggleCourseFavorite(ctx, db.ToggleCourseFavoriteParams{
		UserID:   utils.UUID(viewerID),
		CourseID: utils.UUID(courseID),
	})
	if err != nil {
		return ToggleFavoriteResult{}, fmt.Errorf("ToggleCourseFavorite: %w", err)
	}
	return ToggleFavoriteResult{
		Favorited:   row.Favorited,
		FavoritedAt: utils.TimestamptzPtr(row.FavoritedAt),
	}, nil
}

// Service implements the GET /api/me/favorites business logic.
type Service struct {
	repo Repository
}

// NewService wires a favorites Service over the given repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListFavorites returns a page of the viewer's favorited entities
// (ASK-151).
//
// Strategy:
//   - When EntityType is non-nil: query only that one favorites
//     table with OFFSET/LIMIT directly. Simpler + faster path.
//   - When EntityType is nil: fan out to all three queries with
//     LIMIT (offset+limit+1), merge in Go, sort by
//     FavoritedAt DESC with EntityID tiebreak, slice
//     [offset:offset+limit] for the page. The +1 row distinguishes
//     "exactly limit items remain" from "more items exist".
//
// The per-table LIMIT (offset+limit+1) is correct for the merge:
// for any item that lands in positions 0..(offset+limit) of the
// global ordering, no excluded row from any table can be newer
// (because that table already returned its top (offset+limit+1)
// rows). MaxOffset=1000 caps the worst case at 1101 rows per
// table * 3 tables = ~3.3k rows in memory.
//
// Limit + Cursor + EntityType validation is defense in depth --
// the openapi wrapper enforces these at the HTTP boundary; the
// service re-validates so internal Go callers can't bypass it.
func (s *Service) ListFavorites(ctx context.Context, p ListFavoritesParams) (ListFavoritesResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit < MinLimit || limit > MaxLimit {
		return ListFavoritesResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"limit": fmt.Sprintf("must be between %d and %d", MinLimit, MaxLimit),
		})
	}

	if p.EntityType != nil && !p.EntityType.Valid() {
		return ListFavoritesResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"entity_type": "must be 'file', 'study_guide', or 'course'",
		})
	}

	var offset int32
	if p.Cursor != nil && *p.Cursor != "" {
		decoded, err := DecodeCursor(*p.Cursor)
		if err != nil {
			return ListFavoritesResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			})
		}
		offset = decoded
	}

	viewerPgxID := utils.UUID(p.ViewerID)

	if p.EntityType != nil {
		return s.listSingleType(ctx, *p.EntityType, viewerPgxID, offset, limit)
	}

	// Merged path: fan out to all three queries with the +1 trick
	// applied per-table. Each query starts at offset 0 -- the merge
	// + Go-side slice handles the offset, because per-table OFFSET
	// would be wrong (a row at offset 10 in the file table is not
	// at global offset 10).
	perTableLimit := offset + limit + 1
	fileRows, err := s.repo.ListFileFavorites(ctx, db.ListFileFavoritesParams{
		ViewerID:   viewerPgxID,
		PageLimit:  perTableLimit,
		PageOffset: 0,
	})
	if err != nil {
		return ListFavoritesResult{}, fmt.Errorf("ListFavorites: files: %w", err)
	}
	guideRows, err := s.repo.ListStudyGuideFavorites(ctx, db.ListStudyGuideFavoritesParams{
		ViewerID:   viewerPgxID,
		PageLimit:  perTableLimit,
		PageOffset: 0,
	})
	if err != nil {
		return ListFavoritesResult{}, fmt.Errorf("ListFavorites: study guides: %w", err)
	}
	courseRows, err := s.repo.ListCourseFavorites(ctx, db.ListCourseFavoritesParams{
		ViewerID:   viewerPgxID,
		PageLimit:  perTableLimit,
		PageOffset: 0,
	})
	if err != nil {
		return ListFavoritesResult{}, fmt.Errorf("ListFavorites: courses: %w", err)
	}

	merged := make([]FavoriteItem, 0, len(fileRows)+len(guideRows)+len(courseRows))
	for _, r := range fileRows {
		item, err := mapFavoriteFile(r)
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
		}
		merged = append(merged, item)
	}
	for _, r := range guideRows {
		item, err := mapFavoriteStudyGuide(r)
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
		}
		merged = append(merged, item)
	}
	for _, r := range courseRows {
		item, err := mapFavoriteCourse(r)
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
		}
		merged = append(merged, item)
	}

	// FavoritedAt DESC, then EntityID lex ASC for deterministic
	// tie-break. EntityID.String() formats the canonical 8-4-4-4-12
	// lower-case hex form which is byte-stable for comparison.
	sort.SliceStable(merged, func(i, j int) bool {
		if !merged[i].FavoritedAt.Equal(merged[j].FavoritedAt) {
			return merged[i].FavoritedAt.After(merged[j].FavoritedAt)
		}
		return merged[i].EntityID.String() < merged[j].EntityID.String()
	})

	return paginate(merged, offset, limit), nil
}

// listSingleType handles the EntityType-filter case where one
// per-table query gives us the page directly via SQL OFFSET/LIMIT.
// Mapping uses the same +1 trick for has_more detection.
func (s *Service) listSingleType(ctx context.Context, et EntityType, viewer pgtype.UUID, offset, limit int32) (ListFavoritesResult, error) {
	switch et {
	case EntityTypeFile:
		rows, err := s.repo.ListFileFavorites(ctx, db.ListFileFavoritesParams{
			ViewerID:   viewer,
			PageLimit:  limit + 1,
			PageOffset: offset,
		})
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: files: %w", err)
		}
		out := make([]FavoriteItem, 0, len(rows))
		for _, r := range rows {
			item, err := mapFavoriteFile(r)
			if err != nil {
				return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
			}
			out = append(out, item)
		}
		return paginateSingle(out, offset, limit), nil

	case EntityTypeStudyGuide:
		rows, err := s.repo.ListStudyGuideFavorites(ctx, db.ListStudyGuideFavoritesParams{
			ViewerID:   viewer,
			PageLimit:  limit + 1,
			PageOffset: offset,
		})
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: study guides: %w", err)
		}
		out := make([]FavoriteItem, 0, len(rows))
		for _, r := range rows {
			item, err := mapFavoriteStudyGuide(r)
			if err != nil {
				return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
			}
			out = append(out, item)
		}
		return paginateSingle(out, offset, limit), nil

	case EntityTypeCourse:
		rows, err := s.repo.ListCourseFavorites(ctx, db.ListCourseFavoritesParams{
			ViewerID:   viewer,
			PageLimit:  limit + 1,
			PageOffset: offset,
		})
		if err != nil {
			return ListFavoritesResult{}, fmt.Errorf("ListFavorites: courses: %w", err)
		}
		out := make([]FavoriteItem, 0, len(rows))
		for _, r := range rows {
			item, err := mapFavoriteCourse(r)
			if err != nil {
				return ListFavoritesResult{}, fmt.Errorf("ListFavorites: %w", err)
			}
			out = append(out, item)
		}
		return paginateSingle(out, offset, limit), nil
	}
	// Unreachable: et.Valid() in ListFavorites already gated this.
	return ListFavoritesResult{}, fmt.Errorf("ListFavorites: unknown entity type: %s", et)
}

// paginate slices the post-merge sorted result to the requested
// [offset:offset+limit] window. has_more is true when the source
// slice has at least one row past the window. Used by the merged
// (no entity_type filter) path.
func paginate(merged []FavoriteItem, offset, limit int32) ListFavoritesResult {
	total := int32(len(merged))
	if offset >= total {
		return ListFavoritesResult{
			Favorites:  []FavoriteItem{},
			HasMore:    false,
			NextCursor: nil,
		}
	}
	end := offset + limit
	hasMore := total > end
	if end > total {
		end = total
	}
	// Defensive copy so the returned slice does not pin the
	// merged backing array (the caller's struct may outlive this
	// function and we want the GC to reclaim non-page rows).
	page := append([]FavoriteItem(nil), merged[offset:end]...)
	out := ListFavoritesResult{
		Favorites: page,
		HasMore:   hasMore,
	}
	if hasMore {
		next := EncodeCursor(end)
		out.NextCursor = &next
	}
	return out
}

// paginateSingle handles the single-type case where SQL already
// applied OFFSET/LIMIT and returned at most (limit+1) rows. The
// (limit+1)-th row signals has_more; we trim to limit before
// returning.
func paginateSingle(out []FavoriteItem, offset, limit int32) ListFavoritesResult {
	hasMore := int32(len(out)) > limit
	if hasMore {
		out = out[:limit]
	}
	res := ListFavoritesResult{
		Favorites: out,
		HasMore:   hasMore,
	}
	if hasMore {
		next := EncodeCursor(offset + limit)
		res.NextCursor = &next
	}
	return res
}
