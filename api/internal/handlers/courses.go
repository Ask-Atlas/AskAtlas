package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CourseService defines the application logic required by the CoursesHandler.
type CourseService interface {
	ListCourses(ctx context.Context, params courses.ListCoursesParams) (courses.ListCoursesResult, error)
	GetCourse(ctx context.Context, params courses.GetCourseParams) (courses.CourseDetail, error)
}

// CoursesHandler manages incoming HTTP requests relating to course operations.
type CoursesHandler struct {
	service CourseService
}

// NewCoursesHandler creates a new CoursesHandler backed by the given CourseService.
func NewCoursesHandler(service CourseService) *CoursesHandler {
	return &CoursesHandler{service: service}
}

// ListCourses handles GET /courses, returning a paginated list of courses
// (with embedded school summaries) optionally filtered by school_id,
// department, or a q search term.
func (h *CoursesHandler) ListCourses(w http.ResponseWriter, r *http.Request, params api.ListCoursesParams) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var cursor *courses.Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		decoded, err := courses.DecodeCursor(*params.Cursor)
		if err != nil {
			apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			}))
			return
		}
		cursor = &decoded
	}

	var schoolID *uuid.UUID
	if params.SchoolId != nil {
		id := uuid.UUID(*params.SchoolId)
		schoolID = &id
	}

	var sortBy courses.SortField
	if params.SortBy != nil {
		sortBy = courses.SortField(*params.SortBy)
	}
	var sortDir courses.SortDir
	if params.SortDir != nil {
		sortDir = courses.SortDir(*params.SortDir)
	}

	// Defense-in-depth: clamp before the int32 narrowing cast so a pathological
	// value can't wrap. The openapi validator caps page_limit at 100 at the
	// HTTP boundary and the service layer also clamps; this is the third line
	// of defense for in-process callers and any future routes bypassing the
	// validator.
	var limit int32
	if params.PageLimit != nil {
		v := *params.PageLimit
		if v > int(courses.MaxPageLimit) {
			v = int(courses.MaxPageLimit)
		}
		limit = int32(v)
	}

	result, err := h.service.ListCourses(r.Context(), courses.ListCoursesParams{
		SchoolID:   schoolID,
		Department: params.Department,
		Q:          params.Q,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Limit:      limit,
		Cursor:     cursor,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListCourses failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListCoursesResponse(result))
}

// GetCourse handles GET /courses/{course_id}, returning a single course
// with its embedded school summary and inline sections array.
func (h *CoursesHandler) GetCourse(w http.ResponseWriter, r *http.Request, courseID openapi_types.UUID) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	detail, err := h.service.GetCourse(r.Context(), courses.GetCourseParams{
		CourseID: uuid.UUID(courseID),
	})
	if err != nil {
		// Translate the generic ErrNotFound into a course-specific 404 message
		// before falling back to the standard ToHTTPError mapping (which would
		// otherwise return "Resource not found").
		if errors.Is(err, apperrors.ErrNotFound) {
			apperrors.RespondWithError(w, apperrors.NewNotFound("Course not found"))
			return
		}
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("GetCourse failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapCourseDetailResponse(detail))
}

// JoinSection is a temporary stub satisfying the generated ServerInterface
// while the membership wiring lands in follow-up commits.
func (h *CoursesHandler) JoinSection(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ openapi_types.UUID) {
	apperrors.RespondWithError(w, &apperrors.AppError{
		Code:    http.StatusNotImplemented,
		Status:  "Not Implemented",
		Message: "Endpoint not yet implemented",
	})
}

// LeaveSection is a temporary stub satisfying the generated ServerInterface
// while the membership wiring lands in follow-up commits.
func (h *CoursesHandler) LeaveSection(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ openapi_types.UUID) {
	apperrors.RespondWithError(w, &apperrors.AppError{
		Code:    http.StatusNotImplemented,
		Status:  "Not Implemented",
		Message: "Endpoint not yet implemented",
	})
}

// mapSchoolSummary projects the embedded school summary to its wire shape.
func mapSchoolSummary(s courses.SchoolSummary) api.SchoolSummary {
	return api.SchoolSummary{
		Id:      openapi_types.UUID(s.ID),
		Name:    s.Name,
		Acronym: s.Acronym,
		City:    s.City,
		State:   s.State,
		Country: s.Country,
	}
}

// mapCourseResponse projects a single Course to the list-row wire shape.
func mapCourseResponse(c courses.Course) api.CourseResponse {
	return api.CourseResponse{
		Id:          openapi_types.UUID(c.ID),
		School:      mapSchoolSummary(c.School),
		Department:  c.Department,
		Number:      c.Number,
		Title:       c.Title,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

// mapListCoursesResponse projects the domain ListCoursesResult to the
// generated paginated wire response.
func mapListCoursesResponse(r courses.ListCoursesResult) api.ListCoursesResponse {
	out := make([]api.CourseResponse, 0, len(r.Courses))
	for _, c := range r.Courses {
		out = append(out, mapCourseResponse(c))
	}
	return api.ListCoursesResponse{
		Courses:    out,
		HasMore:    r.HasMore,
		NextCursor: r.NextCursor,
	}
}

// mapSectionSummary projects a single Section to its wire shape.
func mapSectionSummary(s courses.Section) api.SectionSummary {
	return api.SectionSummary{
		Id:             openapi_types.UUID(s.ID),
		Term:           s.Term,
		SectionCode:    s.SectionCode,
		InstructorName: s.InstructorName,
		MemberCount:    s.MemberCount,
	}
}

// mapCourseDetailResponse projects the domain CourseDetail (course + school +
// sections) to the get-by-id wire response. sections is always non-nil so the
// JSON output is "sections": [] for courses with no sections, matching the spec.
func mapCourseDetailResponse(d courses.CourseDetail) api.CourseDetailResponse {
	sections := make([]api.SectionSummary, 0, len(d.Sections))
	for _, s := range d.Sections {
		sections = append(sections, mapSectionSummary(s))
	}
	return api.CourseDetailResponse{
		Id:          openapi_types.UUID(d.ID),
		School:      mapSchoolSummary(d.School),
		Department:  d.Department,
		Number:      d.Number,
		Title:       d.Title,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		Sections:    sections,
	}
}
