package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	ListMyEnrollments(ctx context.Context, params courses.ListMyEnrollmentsParams) ([]courses.Enrollment, error)
	CheckMembership(ctx context.Context, params courses.CheckMembershipParams) (courses.MembershipCheck, error)
	ListSectionMembers(ctx context.Context, params courses.ListSectionMembersParams) (courses.ListSectionMembersResult, error)
	ListCourseSections(ctx context.Context, params courses.ListCourseSectionsParams) (courses.ListCourseSectionsResult, error)
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
// The viewer joins the section as a 'student'. Any *fields* in the request
// body are intentionally ignored -- the role is enforced by the service +
// SQL layer to prevent privilege escalation via {"role": "instructor"} --
// but malformed JSON is still rejected with 400 per the spec, so we
// shallow-parse the body when one is provided.
func (h *CoursesHandler) JoinSection(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, sectionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if appErr := validateJoinSectionBody(r); appErr != nil {
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

// validateJoinSectionBody enforces the ASK-132 input-validation contract
// for the POST body: empty body / missing Content-Type / `{}` / unknown
// fields are all accepted, but a non-empty body that does not parse as
// JSON returns 400 VALIDATION_ERROR. We decode into an empty struct so
// any *fields* present in valid JSON are silently dropped -- the role is
// enforced by the SQL layer regardless.
func validateJoinSectionBody(r *http.Request) *apperrors.AppError {
	if r.Body == nil {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperrors.NewBadRequest("Invalid request body", nil)
	}
	if len(body) == 0 {
		return nil
	}
	var ignored struct{}
	if err := json.Unmarshal(body, &ignored); err != nil {
		return apperrors.NewBadRequest("Invalid request body", nil)
	}
	return nil
}

// ListSectionMembers handles GET /courses/{course_id}/sections/{section_id}/members.
// Any authenticated user can list -- no membership check on the caller
// (course pages are public within the app). The response payload is the
// privacy floor: 5 fields, no email, no clerk_id.
func (h *CoursesHandler) ListSectionMembers(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, sectionId openapi_types.UUID, params api.ListSectionMembersParams) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	svcParams := courses.ListSectionMembersParams{
		CourseID:  uuid.UUID(courseId),
		SectionID: uuid.UUID(sectionId),
	}
	if params.Role != nil {
		role := courses.MemberRole(*params.Role)
		svcParams.Role = &role
	}
	if params.Limit != nil {
		svcParams.Limit = int32(*params.Limit)
	}
	if params.Cursor != nil {
		cur, err := courses.DecodeMemberCursor(*params.Cursor)
		if err != nil {
			apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			}))
			return
		}
		svcParams.Cursor = &cur
	}

	result, err := h.service.ListSectionMembers(r.Context(), svcParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListSectionMembers failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListSectionMembersResponse(result))
}

// ListMyEnrollments handles GET /me/courses, returning every section the
// authenticated viewer is enrolled in. Filters on term + role come from
// the query string; the openapi layer enforces role enum membership and
// term maxLength so the service path-validation is defense in depth.
func (h *CoursesHandler) ListMyEnrollments(w http.ResponseWriter, r *http.Request, params api.ListMyEnrollmentsParams) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	svcParams := courses.ListMyEnrollmentsParams{UserID: viewerID, Term: params.Term}
	if params.Role != nil {
		role := courses.MemberRole(*params.Role)
		svcParams.Role = &role
	}

	enrollments, err := h.service.ListMyEnrollments(r.Context(), svcParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListMyEnrollments failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListMyEnrollmentsResponse(enrollments))
}

// CheckMembership handles GET /courses/{course_id}/sections/{section_id}/members/me.
// Always returns 200 -- non-membership is enrolled=false with null
// role/joined_at, NOT 404. 404 is reserved for missing course/section
// (so the frontend can distinguish "not enrolled" from "section deleted").
func (h *CoursesHandler) CheckMembership(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, sectionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	check, err := h.service.CheckMembership(r.Context(), courses.CheckMembershipParams{
		CourseID:  uuid.UUID(courseId),
		SectionID: uuid.UUID(sectionId),
		UserID:    viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CheckMembership failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapMembershipCheckResponse(check))
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

// mapSectionMemberResponse projects a single SectionMember to the wire
// shape returned by ListSectionMembers. Five fields only -- the schema,
// the SQL, the domain type, AND this mapper all enforce the privacy
// floor (no email, no clerk_id).
func mapSectionMemberResponse(m courses.SectionMember) api.SectionMemberResponse {
	return api.SectionMemberResponse{
		UserId:    openapi_types.UUID(m.UserID),
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Role:      api.SectionMemberResponseRole(m.Role),
		JoinedAt:  m.JoinedAt,
	}
}

// mapListSectionMembersResponse projects the domain result onto the
// paginated wire envelope. members is always non-nil so the JSON output
// is "members": [] rather than null when the section has none.
func mapListSectionMembersResponse(r courses.ListSectionMembersResult) api.ListSectionMembersResponse {
	out := make([]api.SectionMemberResponse, 0, len(r.Members))
	for _, m := range r.Members {
		out = append(out, mapSectionMemberResponse(m))
	}
	return api.ListSectionMembersResponse{
		Members:    out,
		HasMore:    r.HasMore,
		NextCursor: r.NextCursor,
	}
}

// mapEnrollmentResponse projects a single Enrollment to its wire shape.
func mapEnrollmentResponse(e courses.Enrollment) api.EnrollmentResponse {
	return api.EnrollmentResponse{
		Section: api.EnrollmentSectionSummary{
			Id:             openapi_types.UUID(e.Section.ID),
			Term:           e.Section.Term,
			SectionCode:    e.Section.SectionCode,
			InstructorName: e.Section.InstructorName,
		},
		Course: api.EnrollmentCourseSummary{
			Id:         openapi_types.UUID(e.Course.ID),
			Department: e.Course.Department,
			Number:     e.Course.Number,
			Title:      e.Course.Title,
		},
		School: api.EnrollmentSchoolSummary{
			Id:      openapi_types.UUID(e.School.ID),
			Acronym: e.School.Acronym,
		},
		Role:     api.EnrollmentResponseRole(e.Role),
		JoinedAt: e.JoinedAt,
	}
}

// mapListMyEnrollmentsResponse projects the slice of domain Enrollments
// to the wire envelope. Always non-nil so the JSON output is
// "enrollments": [] rather than null when the user has none.
func mapListMyEnrollmentsResponse(enrollments []courses.Enrollment) api.ListMyEnrollmentsResponse {
	out := make([]api.EnrollmentResponse, 0, len(enrollments))
	for _, e := range enrollments {
		out = append(out, mapEnrollmentResponse(e))
	}
	return api.ListMyEnrollmentsResponse{Enrollments: out}
}

// mapMembershipCheckResponse projects MembershipCheck onto the wire
// shape. Role and JoinedAt are pointer types so they marshal as JSON
// null (not omitted) when the viewer is not enrolled, matching the
// schema's nullable: true on both fields.
func mapMembershipCheckResponse(c courses.MembershipCheck) api.MembershipCheckResponse {
	resp := api.MembershipCheckResponse{Enrolled: c.Enrolled}
	if c.Role != nil {
		role := api.MembershipCheckResponseRole(*c.Role)
		resp.Role = &role
	}
	if c.JoinedAt != nil {
		t := *c.JoinedAt
		resp.JoinedAt = &t
	}
	return resp
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

// ListCourseSections handles GET /courses/{course_id}/sections
// (ASK-127). Validates auth, decodes the optional term filter,
// dispatches to the service. The 404 vs 200-empty distinction
// (course missing vs course exists with no matching sections) is
// driven entirely by the service's CourseExists preflight; the
// handler is a thin auth + dispatch + render pass.
//
// Per spec, no GetCourse-style "Course not found" message
// re-mapping is needed -- the service constructs the typed
// 404 with the right message before reaching here.
func (h *CoursesHandler) ListCourseSections(w http.ResponseWriter, r *http.Request, courseID openapi_types.UUID, params api.ListCourseSectionsParams) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	result, err := h.service.ListCourseSections(r.Context(), courses.ListCourseSectionsParams{
		CourseID: uuid.UUID(courseID),
		Term:     params.Term,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListCourseSections failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListCourseSectionsResponse(result))
}

// mapListCourseSectionsResponse projects ListCourseSectionsResult
// onto the wire envelope. Always emits a non-nil Sections slice so
// the JSON output is "sections": [] (not null) on courses with no
// matching sections.
func mapListCourseSectionsResponse(r courses.ListCourseSectionsResult) api.ListCourseSectionsResponse {
	out := make([]api.SectionResponse, 0, len(r.Sections))
	for _, s := range r.Sections {
		out = append(out, mapSectionResponse(s))
	}
	return api.ListCourseSectionsResponse{Sections: out}
}

// mapSectionResponse projects a domain SectionListing onto the
// wire SectionResponse. The two nullable wire fields (section_code,
// instructor_name) are pointer-typed on the domain side, so a nil
// pointer renders as JSON null per the openapi nullable: true
// declaration.
func mapSectionResponse(s courses.SectionListing) api.SectionResponse {
	return api.SectionResponse{
		Id:             openapi_types.UUID(s.ID),
		CourseId:       openapi_types.UUID(s.CourseID),
		Term:           s.Term,
		SectionCode:    s.SectionCode,
		InstructorName: s.InstructorName,
		MemberCount:    s.MemberCount,
		CreatedAt:      s.CreatedAt,
	}
}
