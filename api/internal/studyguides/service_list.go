package studyguides

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// dbFilters holds the resolved shared filter values passed into each
// per-sort-variant query.
type dbFilters struct {
	CourseID pgtype.UUID
	Q        pgtype.Text
	Tags     []string
	Cursor   *Cursor
}

// ListStudyGuides returns a paginated list of study guides for a
// course. The handler runs the AssertCourseExists preflight before
// calling this to map missing courses to 404; by the time we get here,
// the course is known to exist.
//
// Defensive validation mirrors the courses service: limit is clamped
// to [1, MaxPageLimit], sort defaults to score/desc, tags are bounded
// by MaxTagLength, q by MaxSearchLength. oapi-codegen enforces these
// at the wrapper layer in production; the service re-validates so
// direct Go callers can't bypass the contract.
func (s *Service) ListStudyGuides(ctx context.Context, p ListStudyGuidesParams) (ListStudyGuidesResult, error) {
	if err := validateListParams(p); err != nil {
		return ListStudyGuidesResult{}, err
	}

	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	sortBy := p.SortBy
	if sortBy == "" {
		sortBy = SortFieldScore
	}
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDirDesc
	}

	qfn, ok := s.queryTable[sortKey{sortBy, sortDir}]
	if !ok {
		return ListStudyGuidesResult{}, fmt.Errorf("ListStudyGuides: unsupported sort: %s/%s", sortBy, sortDir)
	}

	rows, err := qfn(ctx, toDBFilters(p), limit+1)
	if err != nil {
		return ListStudyGuidesResult{}, fmt.Errorf("ListStudyGuides: %w", err)
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor *string
	if hasMore {
		last := rows[len(rows)-1]
		token, err := EncodeCursor(buildCursor(last, sortBy))
		if err != nil {
			return ListStudyGuidesResult{}, fmt.Errorf("ListStudyGuides: encode cursor: %w", err)
		}
		nextCursor = &token
	}

	return ListStudyGuidesResult{
		StudyGuides: rows,
		HasMore:     hasMore,
		NextCursor:  nextCursor,
	}, nil
}

// validateListParams is the service-layer defensive re-validation.
// Openapi enforces these at the wrapper, but Go callers (including
// tests) could bypass so we re-check here.
func validateListParams(p ListStudyGuidesParams) error {
	if p.Q != nil && len(*p.Q) > MaxSearchLength {
		return apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"q": fmt.Sprintf("must be %d characters or fewer", MaxSearchLength),
		})
	}
	for _, t := range p.Tags {
		if len(t) > MaxTagLength {
			return apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"tag": fmt.Sprintf("each value must be %d characters or fewer", MaxTagLength),
			})
		}
	}
	return nil
}

// toDBFilters resolves the public ListStudyGuidesParams into pgtype
// values shared by every per-sort-variant query.
func toDBFilters(p ListStudyGuidesParams) dbFilters {
	f := dbFilters{
		CourseID: utils.UUID(p.CourseID),
		Cursor:   p.Cursor,
	}
	if p.Q != nil {
		trimmed := strings.TrimSpace(*p.Q)
		if trimmed != "" {
			f.Q = pgtype.Text{String: escapeLikePattern(trimmed), Valid: true}
		}
	}
	if len(p.Tags) > 0 {
		// sqlc emits []string for the tags narg; a nil slice maps to
		// SQL NULL (IS NULL short-circuits the filter).
		f.Tags = append([]string(nil), p.Tags...)
	}
	return f
}

// escapeLikePattern escapes the SQL LIKE/ILIKE wildcards %, _, and \
// so a user-supplied q like "50%_off" is treated as a literal
// substring rather than a wildcard pattern. The SQL queries declare
// ESCAPE '\'. Copy of the courses-package helper; deferred extraction
// to a shared utils package until a third consumer.
func escapeLikePattern(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(s)
}

// buildCursor builds the keyset cursor for the next page from the last
// visible guide row. Only the fields relevant to the active sort are
// populated on the Cursor value; the rest stay nil and are omitted by
// the JSON encoder.
func buildCursor(g StudyGuide, sortBy SortField) Cursor {
	cur := Cursor{ID: g.ID}
	switch sortBy {
	case SortFieldScore:
		vs, vc := g.VoteScore, g.ViewCount
		upd := g.UpdatedAt
		cur.VoteScore = &vs
		cur.ViewCount = &vc
		cur.UpdatedAt = &upd
	case SortFieldViews:
		vc := g.ViewCount
		upd := g.UpdatedAt
		cur.ViewCount = &vc
		cur.UpdatedAt = &upd
	case SortFieldNewest:
		ca := g.CreatedAt
		cur.CreatedAt = &ca
	case SortFieldUpdated:
		upd := g.UpdatedAt
		cur.UpdatedAt = &upd
	}
	return cur
}

// mapListRows projects a slice of typed sqlc rows into domain
// StudyGuides by running the variant-specific row->sharedGuideRow
// adapter and then the shared mapper.
func mapListRows[R any](rows []R, project func(R) sharedGuideRow) ([]StudyGuide, error) {
	out := make([]StudyGuide, 0, len(rows))
	for _, r := range rows {
		g, err := mapStudyGuide(project(r))
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

// Per-sort-variant query methods. Each builds the typed *Params struct,
// calls the matching repository method, and projects the rows through
// the shared mapper.

func (s *Service) queryScoreDesc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesScoreDesc(ctx, db.ListStudyGuidesScoreDescParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorVoteScore: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.VoteScore }),
		CursorViewCount: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromScoreDescRow)
}

func (s *Service) queryScoreAsc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesScoreAsc(ctx, db.ListStudyGuidesScoreAscParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorVoteScore: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.VoteScore }),
		CursorViewCount: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromScoreAscRow)
}

func (s *Service) queryViewsDesc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesViewsDesc(ctx, db.ListStudyGuidesViewsDescParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorViewCount: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromViewsDescRow)
}

func (s *Service) queryViewsAsc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesViewsAsc(ctx, db.ListStudyGuidesViewsAscParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorViewCount: utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromViewsAscRow)
}

func (s *Service) queryNewestDesc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesNewestDesc(ctx, db.ListStudyGuidesNewestDescParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromNewestDescRow)
}

func (s *Service) queryNewestAsc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesNewestAsc(ctx, db.ListStudyGuidesNewestAscParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromNewestAscRow)
}

func (s *Service) queryUpdatedDesc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesUpdatedDesc(ctx, db.ListStudyGuidesUpdatedDescParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromUpdatedDescRow)
}

func (s *Service) queryUpdatedAsc(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error) {
	rows, err := s.repo.ListStudyGuidesUpdatedAsc(ctx, db.ListStudyGuidesUpdatedAscParams{
		CourseID:        f.CourseID,
		Q:               f.Q,
		Tags:            f.Tags,
		PageLimit:       limit,
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromUpdatedAscRow)
}

// ListMyStudyGuides returns the viewer's own study guides (ASK-131).
// Unlike ListStudyGuides, this endpoint INCLUDES soft-deleted rows so
// the owner can see (and eventually restore) their own deleted
// content. The `deleted_at` column surfaces on every row --
// `MyStudyGuide.DeletedAt` is nil for live guides, non-nil for
// soft-deleted ones.
//
// Sort resolution:
//   - MySortFieldUpdated (default) -> ORDER BY updated_at DESC
//   - MySortFieldNewest            -> ORDER BY created_at DESC
//   - MySortFieldTitle             -> ORDER BY LOWER(title) ASC
//
// Unlike ListStudyGuides, there is NO sort_dir knob; each sort
// variant has a single canonical direction per the spec.
func (s *Service) ListMyStudyGuides(ctx context.Context, p ListMyStudyGuidesParams) (ListMyStudyGuidesResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	sortBy := p.SortBy
	if sortBy == "" {
		sortBy = MySortFieldUpdated
	}

	var (
		rows []MyStudyGuide
		err  error
	)
	switch sortBy {
	case MySortFieldUpdated:
		rows, err = s.queryMyUpdated(ctx, p, limit+1)
	case MySortFieldNewest:
		rows, err = s.queryMyNewest(ctx, p, limit+1)
	case MySortFieldTitle:
		rows, err = s.queryMyTitle(ctx, p, limit+1)
	default:
		return ListMyStudyGuidesResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"sort_by": "must be one of: updated, newest, title",
		})
	}
	if err != nil {
		return ListMyStudyGuidesResult{}, fmt.Errorf("ListMyStudyGuides: %w", err)
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor *string
	if hasMore {
		last := rows[len(rows)-1]
		token, encErr := EncodeMyCursor(buildMyCursor(last, sortBy))
		if encErr != nil {
			return ListMyStudyGuidesResult{}, fmt.Errorf("ListMyStudyGuides: encode cursor: %w", encErr)
		}
		nextCursor = &token
	}

	return ListMyStudyGuidesResult{
		StudyGuides: rows,
		HasMore:     hasMore,
		NextCursor:  nextCursor,
	}, nil
}

// buildMyCursor populates only the sort-relevant field on the
// MyCursor so the encoded token carries the minimum state needed to
// advance past the last visible row. ID is always the tiebreaker.
func buildMyCursor(g MyStudyGuide, sortBy MySortField) MyCursor {
	c := MyCursor{ID: g.ID}
	switch sortBy {
	case MySortFieldUpdated:
		t := g.UpdatedAt
		c.UpdatedAt = &t
	case MySortFieldNewest:
		t := g.CreatedAt
		c.CreatedAt = &t
	case MySortFieldTitle:
		lower := strings.ToLower(g.Title)
		c.TitleLower = &lower
	}
	return c
}

// optionalCourseIDPgx converts *uuid.UUID into a pgtype.UUID where
// Valid=false encodes SQL NULL. The ListMyStudyGuides* queries use
// `(sqlc.narg(course_id) IS NULL OR sg.course_id = sqlc.narg(course_id))`
// to short-circuit the filter, so a nil pointer here disables the
// filter entirely.
func optionalCourseIDPgx(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return utils.UUID(*id)
}

// myCursorTimestamptz / myCursorText / myCursorUUID mirror the
// utils.Cursor* helpers but over the MyCursor shape. Kept local so
// the Cursor and MyCursor types can evolve independently.
func myCursorTimestamptz(c *MyCursor, pick func(*MyCursor) *time.Time) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{}
	}
	t := pick(c)
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func myCursorText(c *MyCursor, pick func(*MyCursor) *string) pgtype.Text {
	if c == nil {
		return pgtype.Text{}
	}
	s := pick(c)
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func myCursorUUID(c *MyCursor) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{}
	}
	return utils.UUID(c.ID)
}

func (s *Service) queryMyUpdated(ctx context.Context, p ListMyStudyGuidesParams, limit int32) ([]MyStudyGuide, error) {
	rows, err := s.repo.ListMyStudyGuidesUpdated(ctx, db.ListMyStudyGuidesUpdatedParams{
		CreatorID:       utils.UUID(p.ViewerID),
		CourseID:        optionalCourseIDPgx(p.CourseID),
		PageLimit:       limit,
		CursorUpdatedAt: myCursorTimestamptz(p.Cursor, func(c *MyCursor) *time.Time { return c.UpdatedAt }),
		CursorID:        myCursorUUID(p.Cursor),
	})
	if err != nil {
		return nil, err
	}
	return mapMyListRows(rows, fromMyUpdatedRow)
}

func (s *Service) queryMyNewest(ctx context.Context, p ListMyStudyGuidesParams, limit int32) ([]MyStudyGuide, error) {
	rows, err := s.repo.ListMyStudyGuidesNewest(ctx, db.ListMyStudyGuidesNewestParams{
		CreatorID:       utils.UUID(p.ViewerID),
		CourseID:        optionalCourseIDPgx(p.CourseID),
		PageLimit:       limit,
		CursorCreatedAt: myCursorTimestamptz(p.Cursor, func(c *MyCursor) *time.Time { return c.CreatedAt }),
		CursorID:        myCursorUUID(p.Cursor),
	})
	if err != nil {
		return nil, err
	}
	return mapMyListRows(rows, fromMyNewestRow)
}

func (s *Service) queryMyTitle(ctx context.Context, p ListMyStudyGuidesParams, limit int32) ([]MyStudyGuide, error) {
	rows, err := s.repo.ListMyStudyGuidesTitle(ctx, db.ListMyStudyGuidesTitleParams{
		CreatorID:        utils.UUID(p.ViewerID),
		CourseID:         optionalCourseIDPgx(p.CourseID),
		PageLimit:        limit,
		CursorTitleLower: myCursorText(p.Cursor, func(c *MyCursor) *string { return c.TitleLower }),
		CursorID:         myCursorUUID(p.Cursor),
	})
	if err != nil {
		return nil, err
	}
	return mapMyListRows(rows, fromMyTitleRow)
}

// mapMyListRows mirrors mapListRows but for the MyStudyGuide variants.
// Generic over the sqlc-generated row type so all three query helpers
// share the mapping loop.
func mapMyListRows[R any](rows []R, project func(R) sharedMyGuideRow) ([]MyStudyGuide, error) {
	out := make([]MyStudyGuide, 0, len(rows))
	for _, r := range rows {
		g, err := mapMyStudyGuide(project(r))
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}
