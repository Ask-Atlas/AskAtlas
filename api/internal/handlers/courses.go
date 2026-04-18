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
	JoinSection(ctx context.Context, params courses.JoinSectionParams) (courses.Membership, error)
	LeaveSection(ctx context.Context, params courses.LeaveSectionParams) error
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

// JoinSection handles POST /courses/{course_id}/sections/{section_id}/members.
// The viewer joins the section as a 'student'. Any fields in the request
// body are intentionally ignored; the role is enforced by the service +
// SQL layer to prevent privilege escalation via {"role": "instructor"}.
func (h *CoursesHandler) JoinSection(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, sectionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	membership, err := h.service.JoinSection(r.Context(), courses.JoinSectionParams{
		CourseID:  uuid.UUID(courseId),
		SectionID: uuid.UUID(sectionId),
		UserID:    viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("JoinSection failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapCourseMemberResponse(membership))
}

// LeaveSection handles DELETE /courses/{course_id}/sections/{section_id}/members/me.
// The viewer leaves the section. The /me path segment makes self-only
// scope explicit; there is no path parameter for the user being removed.
func (h *CoursesHandler) LeaveSection(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, sectionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.LeaveSection(r.Context(), courses.LeaveSectionParams{
		CourseID:  uuid.UUID(courseId),
		SectionID: uuid.UUID(sectionId),
		UserID:    viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("LeaveSection failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

// mapCourseMemberResponse projects the domain Membership to the wire shape
// returned by JoinSection (201). The DB enum and the openapi schema enum
// share the same string values so MemberRole maps directly into the
// generated CourseMemberResponseRole type.
func mapCourseMemberResponse(m courses.Membership) api.CourseMemberResponse {
	return api.CourseMemberResponse{
		UserId:    openapi_types.UUID(m.UserID),
		SectionId: openapi_types.UUID(m.SectionID),
		Role:      api.CourseMemberResponseRole(m.Role),
		JoinedAt:  m.JoinedAt,
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
