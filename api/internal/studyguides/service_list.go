package studyguides

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
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

// cursor pgtype helpers unwrap the pointer fields on Cursor into
// pgtype values for the sqlc narg args. Return a zero (invalid)
// pgtype when the cursor is absent or the relevant field is unset,
// which is how the SQL predicates short-circuit to "no cursor applied".

func cursorInt64(c *Cursor, field func(*Cursor) *int64) pgtype.Int8 {
	if c == nil {
		return pgtype.Int8{}
	}
	v := field(c)
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func cursorTimestamp(c *Cursor, field func(*Cursor) *time.Time) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{}
	}
	v := field(c)
	if v == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *v, Valid: true}
}

func cursorID(c *Cursor) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{}
	}
	return utils.UUID(c.ID)
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
		CursorVoteScore: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.VoteScore }),
		CursorViewCount: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorVoteScore: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.VoteScore }),
		CursorViewCount: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorViewCount: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorViewCount: cursorInt64(f.Cursor, func(c *Cursor) *int64 { return c.ViewCount }),
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorCreatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorCreatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
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
		CursorUpdatedAt: cursorTimestamp(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        cursorID(f.Cursor),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromUpdatedAscRow)
}
