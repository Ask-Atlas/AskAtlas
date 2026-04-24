package dashboard

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by the dashboard
// service. The 3 term-resolver methods propagate sql.ErrNoRows
// directly (without mapping to apperrors.ErrNotFound) because the
// service treats "no row" as a control-flow signal -- step 1 falls
// through to step 2, etc. -- not as an end-user error.
type Repository interface {
	// Term-resolution waterfall.
	ResolveCurrentTermActive(ctx context.Context, viewerID pgtype.UUID) (string, error)
	ResolveCurrentTermLastEnded(ctx context.Context, viewerID pgtype.UUID) (string, error)
	ResolveCurrentTermLexLatest(ctx context.Context, viewerID pgtype.UUID) (string, error)

	// Courses section.
	ListEnrolledCoursesForTerm(ctx context.Context, arg db.ListEnrolledCoursesForTermParams) ([]db.ListEnrolledCoursesForTermRow, error)

	// Study-guides section.
	CountUserStudyGuides(ctx context.Context, viewerID pgtype.UUID) (int64, error)
	ListRecentUserStudyGuides(ctx context.Context, arg db.ListRecentUserStudyGuidesParams) ([]db.ListRecentUserStudyGuidesRow, error)

	// Practice section.
	GetUserPracticeStats(ctx context.Context, viewerID pgtype.UUID) (db.GetUserPracticeStatsRow, error)
	CountUserAnsweredQuestions(ctx context.Context, viewerID pgtype.UUID) (int64, error)
	ListRecentUserSessions(ctx context.Context, arg db.ListRecentUserSessionsParams) ([]db.ListRecentUserSessionsRow, error)

	// Files section.
	GetUserFileStats(ctx context.Context, viewerID pgtype.UUID) (db.GetUserFileStatsRow, error)
	ListRecentUserFiles(ctx context.Context, arg db.ListRecentUserFilesParams) ([]db.ListRecentUserFilesRow, error)
}

type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) ResolveCurrentTermActive(ctx context.Context, viewerID pgtype.UUID) (string, error) {
	return r.queries.ResolveCurrentTermActive(ctx, viewerID)
}

func (r *sqlcRepository) ResolveCurrentTermLastEnded(ctx context.Context, viewerID pgtype.UUID) (string, error) {
	return r.queries.ResolveCurrentTermLastEnded(ctx, viewerID)
}

func (r *sqlcRepository) ResolveCurrentTermLexLatest(ctx context.Context, viewerID pgtype.UUID) (string, error) {
	return r.queries.ResolveCurrentTermLexLatest(ctx, viewerID)
}

func (r *sqlcRepository) ListEnrolledCoursesForTerm(ctx context.Context, arg db.ListEnrolledCoursesForTermParams) ([]db.ListEnrolledCoursesForTermRow, error) {
	return r.queries.ListEnrolledCoursesForTerm(ctx, arg)
}

func (r *sqlcRepository) CountUserStudyGuides(ctx context.Context, viewerID pgtype.UUID) (int64, error) {
	return r.queries.CountUserStudyGuides(ctx, viewerID)
}

func (r *sqlcRepository) ListRecentUserStudyGuides(ctx context.Context, arg db.ListRecentUserStudyGuidesParams) ([]db.ListRecentUserStudyGuidesRow, error) {
	return r.queries.ListRecentUserStudyGuides(ctx, arg)
}

func (r *sqlcRepository) GetUserPracticeStats(ctx context.Context, viewerID pgtype.UUID) (db.GetUserPracticeStatsRow, error) {
	return r.queries.GetUserPracticeStats(ctx, viewerID)
}

func (r *sqlcRepository) CountUserAnsweredQuestions(ctx context.Context, viewerID pgtype.UUID) (int64, error) {
	return r.queries.CountUserAnsweredQuestions(ctx, viewerID)
}

func (r *sqlcRepository) ListRecentUserSessions(ctx context.Context, arg db.ListRecentUserSessionsParams) ([]db.ListRecentUserSessionsRow, error) {
	return r.queries.ListRecentUserSessions(ctx, arg)
}

func (r *sqlcRepository) GetUserFileStats(ctx context.Context, viewerID pgtype.UUID) (db.GetUserFileStatsRow, error) {
	return r.queries.GetUserFileStats(ctx, viewerID)
}

func (r *sqlcRepository) ListRecentUserFiles(ctx context.Context, arg db.ListRecentUserFilesParams) ([]db.ListRecentUserFilesRow, error) {
	return r.queries.ListRecentUserFiles(ctx, arg)
}
