package courses

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. The 8 list
// methods correspond to the per-sort-variant sqlc queries; ListCourseSections
// powers the inline sections array in the get-by-id response.
type Repository interface {
	ListCoursesDepartmentAsc(ctx context.Context, arg db.ListCoursesDepartmentAscParams) ([]db.ListCoursesDepartmentAscRow, error)
	ListCoursesDepartmentDesc(ctx context.Context, arg db.ListCoursesDepartmentDescParams) ([]db.ListCoursesDepartmentDescRow, error)
	ListCoursesNumberAsc(ctx context.Context, arg db.ListCoursesNumberAscParams) ([]db.ListCoursesNumberAscRow, error)
	ListCoursesNumberDesc(ctx context.Context, arg db.ListCoursesNumberDescParams) ([]db.ListCoursesNumberDescRow, error)
	ListCoursesTitleAsc(ctx context.Context, arg db.ListCoursesTitleAscParams) ([]db.ListCoursesTitleAscRow, error)
	ListCoursesTitleDesc(ctx context.Context, arg db.ListCoursesTitleDescParams) ([]db.ListCoursesTitleDescRow, error)
	ListCoursesCreatedAtAsc(ctx context.Context, arg db.ListCoursesCreatedAtAscParams) ([]db.ListCoursesCreatedAtAscRow, error)
	ListCoursesCreatedAtDesc(ctx context.Context, arg db.ListCoursesCreatedAtDescParams) ([]db.ListCoursesCreatedAtDescRow, error)

	GetCourse(ctx context.Context, id pgtype.UUID) (db.GetCourseRow, error)
	ListCourseSections(ctx context.Context, courseID pgtype.UUID) ([]db.ListCourseSectionsRow, error)
}

// Service is the business-logic layer for the courses feature.
type Service struct {
	repo Repository
}

// NewService creates a new Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}
