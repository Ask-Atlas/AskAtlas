package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// coursesTestRouter wires the composite handler with mocked file/grant/schools
// services so /courses requests resolve through the same routing the real
// binary uses.
func coursesTestRouter(t *testing.T, ch *handlers.CoursesHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

// ------------------------------------------------------------------------
// ListCourses
// ------------------------------------------------------------------------

func TestCoursesHandler_ListCourses_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/courses", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_ListCourses_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	schoolID := uuid.New()
	desc := "Software engineering fundamentals."
	city := "Pullman"
	country := "US"
	created := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListCourses(mock.Anything, mock.Anything).
		Return(courses.ListCoursesResult{
			Courses: []courses.Course{{
				ID: courseID,
				School: courses.SchoolSummary{
					ID: schoolID, Name: "Washington State University", Acronym: "WSU",
					City: &city, Country: &country,
				},
				Department: "CPTS", Number: "322",
				Title:       "Software Engineering Principles I",
				Description: &desc, CreatedAt: created,
			}},
			HasMore: false,
		}, nil)

	req := authedRequest(t, "/courses")
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListCoursesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Courses, 1)
	assert.Equal(t, courseID, uuid.UUID(resp.Courses[0].Id))
	assert.Equal(t, "CPTS", resp.Courses[0].Department)
	assert.Equal(t, "322", resp.Courses[0].Number)
	assert.Equal(t, "WSU", resp.Courses[0].School.Acronym)
	require.NotNil(t, resp.Courses[0].School.City)
	assert.Equal(t, "Pullman", *resp.Courses[0].School.City)
	assert.False(t, resp.HasMore)
}

func TestCoursesHandler_ListCourses_BadCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	req := authedRequest(t, "/courses?cursor=!!!notbase64")
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	details, ok := body["details"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid cursor value", details["cursor"])

	// Cursor decode failure must short-circuit before the service runs.
	mockSvc.AssertNotCalled(t, "ListCourses", mock.Anything, mock.Anything)
}

func TestCoursesHandler_ListCourses_FiltersAndSortForwarded(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	schoolID := uuid.New()

	mockSvc.EXPECT().
		ListCourses(mock.Anything, mock.MatchedBy(func(p courses.ListCoursesParams) bool {
			return p.SchoolID != nil && *p.SchoolID == schoolID &&
				p.Department != nil && *p.Department == "CPTS" &&
				p.Q != nil && *p.Q == "software" &&
				p.SortBy == courses.SortFieldTitle &&
				p.SortDir == courses.SortDirDesc &&
				p.Limit == 10
		})).
		Return(courses.ListCoursesResult{}, nil)

	url := "/courses?school_id=" + schoolID.String() + "&department=CPTS&q=software&sort_by=title&sort_dir=desc&page_limit=10"
	req := authedRequest(t, url)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCoursesHandler_ListCourses_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListCourses(mock.Anything, mock.Anything).
		Return(courses.ListCoursesResult{}, errors.New("db down"))

	req := authedRequest(t, "/courses")
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ------------------------------------------------------------------------
// GetCourse
// ------------------------------------------------------------------------

func TestCoursesHandler_GetCourse_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/courses/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_GetCourse_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	schoolID := uuid.New()
	sectionID := uuid.New()
	sectionCode := "01"
	instructor := "Dr. Ananth Jillepalli"
	created := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		GetCourse(mock.Anything, mock.MatchedBy(func(p courses.GetCourseParams) bool {
			return p.CourseID == courseID
		})).
		Return(courses.CourseDetail{
			Course: courses.Course{
				ID: courseID,
				School: courses.SchoolSummary{
					ID: schoolID, Name: "Washington State University", Acronym: "WSU",
				},
				Department: "CPTS", Number: "322",
				Title:     "Software Engineering Principles I",
				CreatedAt: created,
			},
			Sections: []courses.Section{{
				ID: sectionID, Term: "Spring 2026",
				SectionCode: &sectionCode, InstructorName: &instructor,
				MemberCount: 34,
			}},
		}, nil)

	req := authedRequest(t, "/courses/"+courseID.String())
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.CourseDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, courseID, uuid.UUID(resp.Id))
	assert.Equal(t, "CPTS", resp.Department)
	assert.Equal(t, "WSU", resp.School.Acronym)
	require.Len(t, resp.Sections, 1)
	assert.Equal(t, "Spring 2026", resp.Sections[0].Term)
	require.NotNil(t, resp.Sections[0].SectionCode)
	assert.Equal(t, "01", *resp.Sections[0].SectionCode)
	assert.Equal(t, int64(34), resp.Sections[0].MemberCount)
}

func TestCoursesHandler_GetCourse_NotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(courses.CourseDetail{}, fmt.Errorf("GetCourse: %w", apperrors.ErrNotFound))

	req := authedRequest(t, "/courses/"+uuid.New().String())
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "Course not found", body["message"])
}

func TestCoursesHandler_GetCourse_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(courses.CourseDetail{}, errors.New("db down"))

	req := authedRequest(t, "/courses/"+uuid.New().String())
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ------------------------------------------------------------------------
// JoinSection (ASK-132)
// ------------------------------------------------------------------------

// authedRequestMethod builds an authenticated request for non-GET verbs
// used by the join/leave tests. authedRequest is GET-only by design;
// putting the method-aware helper here keeps the schools_test.go helper
// small and lets the courses tests own their own request shape.
func authedRequestMethod(t *testing.T, method, target string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	ctx := authctx.WithUserID(req.Context(), uuid.New())
	return req.WithContext(ctx)
}

func TestCoursesHandler_JoinSection_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_JoinSection_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()
	joinedAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.MatchedBy(func(p courses.JoinSectionParams) bool {
			return p.CourseID == courseID && p.SectionID == sectionID
		})).
		Return(courses.Membership{
			UserID:    userID,
			SectionID: sectionID,
			Role:      courses.MemberRoleStudent,
			JoinedAt:  joinedAt,
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", courseID, sectionID)
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.CourseMemberResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, userID, uuid.UUID(resp.UserId))
	assert.Equal(t, sectionID, uuid.UUID(resp.SectionId))
	assert.Equal(t, api.Student, resp.Role)
	assert.True(t, resp.JoinedAt.Equal(joinedAt))
}

// Handler must not decode role from the request body. Even when the body
// supplies {"role":"instructor"}, the service is invoked as written and
// the request is treated identically to {} or empty -- the SQL layer
// hardcodes 'student'. This protects against a privilege-escalation
// vector via crafted JSON.
func TestCoursesHandler_JoinSection_IgnoresUnexpectedBodyFields(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{
			UserID:    userID,
			SectionID: sectionID,
			Role:      courses.MemberRoleStudent,
			JoinedAt:  time.Now().UTC(),
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", courseID, sectionID)
	body := strings.NewReader(`{"role":"instructor"}`)
	req := authedRequestMethod(t, http.MethodPost, url, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.CourseMemberResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, api.Student, resp.Role)
}

func TestCoursesHandler_JoinSection_AlreadyMember(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{}, &apperrors.AppError{
			Code:    http.StatusConflict,
			Status:  "Conflict",
			Message: "Already a member of this section",
		})

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Already a member of this section")
}

func TestCoursesHandler_JoinSection_CourseNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{}, apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Course not found")
}

func TestCoursesHandler_JoinSection_SectionNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{}, apperrors.NewNotFound("Section not found"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Section not found")
}

// The oapi-codegen wrapper validates path UUIDs before calling the
// handler. A malformed course_id should never reach the service mock
// (NewMockCourseService(t) would fail on an unexpected call).
func TestCoursesHandler_JoinSection_BadCourseUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/not-a-uuid/sections/%s/members", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoursesHandler_JoinSection_BadSectionUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/not-a-uuid/members", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoursesHandler_JoinSection_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{}, errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ------------------------------------------------------------------------
// LeaveSection (ASK-138)
// ------------------------------------------------------------------------

func TestCoursesHandler_LeaveSection_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_LeaveSection_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	sectionID := uuid.New()

	mockSvc.EXPECT().
		LeaveSection(mock.Anything, mock.MatchedBy(func(p courses.LeaveSectionParams) bool {
			return p.CourseID == courseID && p.SectionID == sectionID
		})).
		Return(nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", courseID, sectionID)
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCoursesHandler_LeaveSection_NotMember(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		LeaveSection(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Not a member of this section"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Not a member of this section")
}

func TestCoursesHandler_LeaveSection_CourseNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		LeaveSection(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Course not found")
}

func TestCoursesHandler_LeaveSection_BadCourseUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/not-a-uuid/sections/%s/members/me", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoursesHandler_LeaveSection_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		LeaveSection(mock.Anything, mock.Anything).
		Return(errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
