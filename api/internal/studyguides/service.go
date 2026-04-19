package studyguides

// The service layer for this package is split across four files:
//
//   - service.go       (this file) -- Repository interface, Service
//                                     struct, constructor.
//   - service_list.go  -- ListStudyGuides + the 8 per-sort-variant
//                         query methods + cursor/filter helpers.
//   - service_read.go  -- GetStudyGuide (parallel sibling-query
//                         hydration) + AssertCourseExists.
//   - service_write.go -- CreateStudyGuide, DeleteStudyGuide,
//                         CastVote, RemoveVote, RecommendStudyGuide,
//                         RemoveRecommendation + write-side helpers.
//
// The canonical `// Package studyguides ...` doc comment lives in
// model.go (Go expects exactly one package-level doc per package).

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
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

	GuideExistsAndLive(ctx context.Context, id pgtype.UUID) (bool, error)
	UpsertStudyGuideVote(ctx context.Context, arg db.UpsertStudyGuideVoteParams) error
	ComputeGuideVoteScore(ctx context.Context, studyGuideID pgtype.UUID) (int64, error)
	DeleteStudyGuideVote(ctx context.Context, arg db.DeleteStudyGuideVoteParams) (int64, error)

	ViewerCanRecommendForGuide(ctx context.Context, arg db.ViewerCanRecommendForGuideParams) (db.ViewerCanRecommendForGuideRow, error)
	InsertStudyGuideRecommendation(ctx context.Context, arg db.InsertStudyGuideRecommendationParams) (db.InsertStudyGuideRecommendationRow, error)
	DeleteStudyGuideRecommendation(ctx context.Context, arg db.DeleteStudyGuideRecommendationParams) (int64, error)

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
