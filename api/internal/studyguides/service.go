package studyguides

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// Repository is the data-access surface required by Service. The 8 list
// methods correspond to the per-sort-variant sqlc queries -- sqlc cannot
// parameterize ORDER BY, so each (sort_by, sort_dir) combination has
// its own typed query. CourseExistsForGuides is the existence probe
// used by the handler to produce a 404 when the course is missing.
type Repository interface {
	ListStudyGuidesScoreDesc(ctx context.Context, arg db.ListStudyGuidesScoreDescParams) ([]db.ListStudyGuidesScoreDescRow, error)
	ListStudyGuidesScoreAsc(ctx context.Context, arg db.ListStudyGuidesScoreAscParams) ([]db.ListStudyGuidesScoreAscRow, error)
	ListStudyGuidesViewsDesc(ctx context.Context, arg db.ListStudyGuidesViewsDescParams) ([]db.ListStudyGuidesViewsDescRow, error)
	ListStudyGuidesViewsAsc(ctx context.Context, arg db.ListStudyGuidesViewsAscParams) ([]db.ListStudyGuidesViewsAscRow, error)
	ListStudyGuidesNewestDesc(ctx context.Context, arg db.ListStudyGuidesNewestDescParams) ([]db.ListStudyGuidesNewestDescRow, error)
	ListStudyGuidesNewestAsc(ctx context.Context, arg db.ListStudyGuidesNewestAscParams) ([]db.ListStudyGuidesNewestAscRow, error)
	ListStudyGuidesUpdatedDesc(ctx context.Context, arg db.ListStudyGuidesUpdatedDescParams) ([]db.ListStudyGuidesUpdatedDescRow, error)
	ListStudyGuidesUpdatedAsc(ctx context.Context, arg db.ListStudyGuidesUpdatedAscParams) ([]db.ListStudyGuidesUpdatedAscRow, error)

	CourseExistsForGuides(ctx context.Context, id pgtype.UUID) (bool, error)

	GetStudyGuideDetail(ctx context.Context, id pgtype.UUID) (db.GetStudyGuideDetailRow, error)
	GetUserVoteForGuide(ctx context.Context, arg db.GetUserVoteForGuideParams) (db.VoteDirection, error)
	ListGuideRecommenders(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideRecommendersRow, error)
	ListGuideQuizzesWithQuestionCount(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideQuizzesWithQuestionCountRow, error)
	ListGuideResources(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideResourcesRow, error)
	ListGuideFiles(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideFilesRow, error)

	InsertStudyGuide(ctx context.Context, arg db.InsertStudyGuideParams) (db.InsertStudyGuideRow, error)
	GetStudyGuideByIDForUpdate(ctx context.Context, id pgtype.UUID) (db.GetStudyGuideByIDForUpdateRow, error)
	SoftDeleteStudyGuide(ctx context.Context, id pgtype.UUID) error
	SoftDeleteQuizzesForGuide(ctx context.Context, studyGuideID pgtype.UUID) error

	// InTx runs fn inside a single Postgres transaction. The Repository
	// passed to fn is scoped to the tx; commits on a nil return,
	// rolls back on any error. Used by DeleteStudyGuide for the
	// atomic guide + child-quiz cascade.
	InTx(ctx context.Context, fn func(Repository) error) error
}

// sortKey is the lookup key for the per-sort-variant query function
// table.
type sortKey struct {
	Field SortField
	Dir   SortDir
}

// queryFn is the signature shared by every per-sort-variant query
// method on Service. It returns already-mapped domain StudyGuides so
// the dispatch site stays variant-agnostic.
type queryFn func(ctx context.Context, f dbFilters, limit int32) ([]StudyGuide, error)

// Service is the business-logic layer for the study-guides feature.
type Service struct {
	repo       Repository
	queryTable map[sortKey]queryFn
}

// NewService creates a new Service backed by the given Repository. The
// queryTable is built once at construction so ListStudyGuides can
// dispatch by sort key with no per-request reflection or type switching.
func NewService(repo Repository) *Service {
	s := &Service{repo: repo}
	s.queryTable = map[sortKey]queryFn{
		{SortFieldScore, SortDirDesc}:   s.queryScoreDesc,
		{SortFieldScore, SortDirAsc}:    s.queryScoreAsc,
		{SortFieldViews, SortDirDesc}:   s.queryViewsDesc,
		{SortFieldViews, SortDirAsc}:    s.queryViewsAsc,
		{SortFieldNewest, SortDirDesc}:  s.queryNewestDesc,
		{SortFieldNewest, SortDirAsc}:   s.queryNewestAsc,
		{SortFieldUpdated, SortDirDesc}: s.queryUpdatedDesc,
		{SortFieldUpdated, SortDirAsc}:  s.queryUpdatedAsc,
	}
	return s
}

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

// AssertCourseExists is the 404-distinguishing preflight. Handler calls
// it before ListStudyGuides to emit a tailored "Course not found" 404
// when the course_id doesn't resolve. Pulling the check here (rather
// than interpreting an empty result) matches the pattern from
// courses.Service.assertCourseAndSection.
func (s *Service) AssertCourseExists(ctx context.Context, courseID uuid.UUID) error {
	exists, err := s.repo.CourseExistsForGuides(ctx, utils.UUID(courseID))
	if err != nil {
		return fmt.Errorf("AssertCourseExists: %w", err)
	}
	if !exists {
		return apperrors.NewNotFound("Course not found")
	}
	return nil
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

// GetStudyGuide fetches the full study-guide detail including nested
// course + creator + recommenders + quizzes + resources + files + the
// viewer's own vote state.
//
// This is a pure read -- no view-count increment, no last-viewed
// upsert, no mutation. View tracking will live on its own dedicated
// POST (future ticket mirroring ASK-134's POST /files/{id}/view).
//
// Runs 6 queries total: GetStudyGuideDetail (404 candidate, must
// complete before the rest fan out so we don't make 5 speculative
// round trips for a guide that doesn't exist), then the 5 independent
// sibling queries (user_vote, recommenders, quizzes, resources,
// files) issued in parallel via errgroup. The 5-wide fan-out cuts
// end-to-end latency from sum(query_time) to max(query_time) while
// preserving the strict error-propagation semantics -- any sibling
// error short-circuits the rest via ctx cancellation and returns a
// 500 to the handler.
//
// Soft-deleted guides return apperrors.ErrNotFound via the underlying
// query's WHERE sg.deleted_at IS NULL. The handler maps that to a
// 404 'Study guide not found'.
func (s *Service) GetStudyGuide(ctx context.Context, p GetStudyGuideParams) (StudyGuideDetail, error) {
	guidePgxID := utils.UUID(p.StudyGuideID)
	viewerPgxID := utils.UUID(p.ViewerID)

	row, err := s.repo.GetStudyGuideDetail(ctx, guidePgxID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StudyGuideDetail{}, apperrors.NewNotFound("Study guide not found")
		}
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: detail: %w", err)
	}

	detail, err := mapStudyGuideDetail(row)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: map detail: %w", err)
	}

	// The 5 sibling queries are independent of each other -- fan out
	// in parallel via errgroup. Each goroutine owns its slot on
	// `detail` (different field per slot), so there's no data race on
	// the struct.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		vote, err := s.repo.GetUserVoteForGuide(gctx, db.GetUserVoteForGuideParams{
			StudyGuideID: guidePgxID,
			ViewerID:     viewerPgxID,
		})
		switch {
		case err == nil:
			v := GuideVote(vote)
			detail.UserVote = &v
		case errors.Is(err, sql.ErrNoRows):
			detail.UserVote = nil
		default:
			return fmt.Errorf("user vote: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideRecommenders(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("recommenders: %w", err)
		}
		detail.RecommendedBy = make([]Creator, 0, len(rows))
		for _, r := range rows {
			rec, err := mapRecommender(r)
			if err != nil {
				return fmt.Errorf("map recommender: %w", err)
			}
			detail.RecommendedBy = append(detail.RecommendedBy, rec)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideQuizzesWithQuestionCount(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("quizzes: %w", err)
		}
		detail.Quizzes = make([]Quiz, 0, len(rows))
		for _, q := range rows {
			quiz, err := mapQuiz(q)
			if err != nil {
				return fmt.Errorf("map quiz: %w", err)
			}
			detail.Quizzes = append(detail.Quizzes, quiz)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideResources(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("resources: %w", err)
		}
		detail.Resources = make([]Resource, 0, len(rows))
		for _, r := range rows {
			res, err := mapResource(r)
			if err != nil {
				return fmt.Errorf("map resource: %w", err)
			}
			detail.Resources = append(detail.Resources, res)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideFiles(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("files: %w", err)
		}
		detail.Files = make([]GuideFile, 0, len(rows))
		for _, f := range rows {
			gf, err := mapGuideFile(f)
			if err != nil {
				return fmt.Errorf("map file: %w", err)
			}
			detail.Files = append(detail.Files, gf)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: %w", err)
	}

	return detail, nil
}

// normalizeTags trims, lowercases, and dedupes a raw input tag slice.
// Returns the cleaned slice or apperrors.NewBadRequest with a per-field
// detail when an entry is empty after trim, exceeds MaxTagLength, or
// the input count exceeds MaxTagsCount.
//
// Always returns a non-nil slice (possibly length 0) so callers can
// rely on the result being safe to pass to the Postgres NOT NULL
// text[] column (study_guides.tags has DEFAULT '{}').
func normalizeTags(in []string) ([]string, error) {
	// Cap on the RAW input count (pre-dedupe), mirroring the openapi
	// schema's `tags.maxItems: 20`. Keep the check in this position --
	// moving it after dedupe would diverge from the schema and let a
	// 1000-item request waste CPU on the loop before being rejected.
	if len(in) > MaxTagsCount {
		return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
			"tags": fmt.Sprintf("must contain %d items or fewer", MaxTagsCount),
		})
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		t := strings.ToLower(strings.TrimSpace(raw))
		if t == "" {
			return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
				"tags": "values must not be empty",
			})
		}
		if len(t) > MaxTagLength {
			return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
				"tags": fmt.Sprintf("each value must be %d characters or fewer", MaxTagLength),
			})
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out, nil
}

// trimmedNonEmpty returns nil if s is nil or trims to empty; otherwise
// returns a pointer to the trimmed string. Used by CreateStudyGuide to
// normalize the optional description + content fields so a body like
// `{"description": "   "}` lands as SQL NULL rather than persisting a
// whitespace-only string.
func trimmedNonEmpty(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

// validateCreateParams runs the service-layer defensive re-validation
// for CreateStudyGuide. openapi enforces these at the wrapper layer in
// production, but Go callers (including tests) could bypass so we
// re-check here. Mirrors validateListParams.
func validateCreateParams(p CreateStudyGuideParams) error {
	if strings.TrimSpace(p.Title) == "" {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": "must not be empty",
		})
	}
	if len(p.Title) > MaxTitleLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": fmt.Sprintf("must be %d characters or fewer", MaxTitleLength),
		})
	}
	if p.Description != nil && len(*p.Description) > MaxDescriptionLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"description": fmt.Sprintf("must be %d characters or fewer", MaxDescriptionLength),
		})
	}
	if p.Content != nil && len(*p.Content) > MaxContentLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"content": fmt.Sprintf("must be %d characters or fewer", MaxContentLength),
		})
	}
	return nil
}

// CreateStudyGuide creates a new study guide owned by the authenticated
// user. Runs AssertCourseExists so a missing course surfaces as a
// tailored 404 (rather than a generic FK-violation 500). After the
// insert, hydrates the response via GetStudyGuideDetail so vote_score
// and is_recommended come out of the same SQL projection used by GET
// /study-guides/{id} -- the privacy floor stays in one place.
//
// The 5 sibling queries from GetStudyGuide (recommenders, quizzes,
// resources, files, user_vote) are intentionally skipped: a freshly
// inserted guide has no children, no votes, and no recommenders by
// definition. The corresponding response slices are emitted as empty
// (non-nil) so the JSON wire shape is `[]` and not `null`.
func (s *Service) CreateStudyGuide(ctx context.Context, p CreateStudyGuideParams) (StudyGuideDetail, error) {
	if err := validateCreateParams(p); err != nil {
		return StudyGuideDetail{}, err
	}

	tags, err := normalizeTags(p.Tags)
	if err != nil {
		return StudyGuideDetail{}, err
	}

	if err := s.AssertCourseExists(ctx, p.CourseID); err != nil {
		return StudyGuideDetail{}, err
	}

	// Title is required; description/content are optional. For all three,
	// trim leading/trailing whitespace and treat the empty result as the
	// absent value so the DB stores SQL NULL (not a whitespace-only
	// string). Title's "absent" form is rejected upstream by
	// validateCreateParams; description/content are dropped to nil.
	inserted, err := s.repo.InsertStudyGuide(ctx, db.InsertStudyGuideParams{
		CourseID:    utils.UUID(p.CourseID),
		CreatorID:   utils.UUID(p.CreatorID),
		Title:       strings.TrimSpace(p.Title),
		Description: utils.Text(trimmedNonEmpty(p.Description)),
		Content:     utils.Text(trimmedNonEmpty(p.Content)),
		Tags:        tags,
	})
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: insert: %w", err)
	}

	row, err := s.repo.GetStudyGuideDetail(ctx, inserted.ID)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: hydrate: %w", err)
	}

	detail, err := mapStudyGuideDetail(row)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: map detail: %w", err)
	}

	detail.UserVote = nil
	detail.RecommendedBy = []Creator{}
	detail.Quizzes = []Quiz{}
	detail.Resources = []Resource{}
	detail.Files = []GuideFile{}

	return detail, nil
}

// DeleteStudyGuide soft-deletes a study guide (creator-only). Wraps
// the locked SELECT + soft-delete + child-quiz cascade in a single
// transaction via repo.InTx so the cascade is atomic: either both the
// guide and its quizzes get deleted_at set, or neither does. The
// SELECT FOR UPDATE in GetStudyGuideByIDForUpdate prevents two
// concurrent deletes from racing on the same guide -- one wins with
// 204, the other sees the row already-deleted in its tx snapshot and
// returns 404.
//
// 404 is returned both when the guide is missing and when it's already
// soft-deleted (idempotent semantics -- a duplicate DELETE shouldn't
// surface a 409 since the desired state is reached). 403 is returned
// when the viewer is not the guide's creator.
func (s *Service) DeleteStudyGuide(ctx context.Context, p DeleteStudyGuideParams) error {
	guidePgxID := utils.UUID(p.StudyGuideID)
	return s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetStudyGuideByIDForUpdate(ctx, guidePgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Study guide not found")
			}
			return fmt.Errorf("DeleteStudyGuide: lock: %w", err)
		}
		if row.DeletedAt.Valid {
			return apperrors.NewNotFound("Study guide not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("DeleteStudyGuide: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}
		if err := tx.SoftDeleteStudyGuide(ctx, guidePgxID); err != nil {
			return fmt.Errorf("DeleteStudyGuide: soft delete guide: %w", err)
		}
		if err := tx.SoftDeleteQuizzesForGuide(ctx, guidePgxID); err != nil {
			return fmt.Errorf("DeleteStudyGuide: soft delete quizzes: %w", err)
		}
		return nil
	})
}
