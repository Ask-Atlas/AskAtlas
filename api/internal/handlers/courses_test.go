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
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, nil, nil, nil)
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
	assert.Equal(t, api.CourseMemberResponseRoleStudent, resp.Role)
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
	assert.Equal(t, api.CourseMemberResponseRoleStudent, resp.Role)
}

// Per ASK-132 input-validation table: malformed JSON in the body must
// surface as 400 even though valid JSON with unexpected fields is
// accepted. NewMockCourseService(t) would fail on an unexpected call,
// proving the handler rejects before reaching the service.
func TestCoursesHandler_JoinSection_MalformedJSONBody(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	body := strings.NewReader(`{invalid`)
	req := authedRequestMethod(t, http.MethodPost, url, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

// Empty body is the documented default and must succeed -- regression
// guard for the malformed-JSON fix above so the body-decode path doesn't
// accidentally reject the documented happy case.
func TestCoursesHandler_JoinSection_EmptyBodyAccepted(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(courses.Membership{
			UserID:    uuid.New(),
			SectionID: uuid.New(),
			Role:      courses.MemberRoleStudent,
			JoinedAt:  time.Now().UTC(),
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
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

// ------------------------------------------------------------------------
// ListMyEnrollments (ASK-154)
// ------------------------------------------------------------------------

func TestCoursesHandler_ListMyEnrollments_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/me/courses", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_ListMyEnrollments_Empty(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.Anything).
		Return(nil, nil)

	req := authedRequestMethod(t, http.MethodGet, "/me/courses", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Envelope must emit [] not null when empty.
	var resp api.ListMyEnrollmentsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotNil(t, resp.Enrollments)
	assert.Empty(t, resp.Enrollments)
}

func TestCoursesHandler_ListMyEnrollments_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	sectionID := uuid.New()
	courseID := uuid.New()
	schoolID := uuid.New()
	code := "01"
	instr := "Dr. Test"
	joinedAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.Anything).
		Return([]courses.Enrollment{{
			Section:  courses.EnrollmentSection{ID: sectionID, Term: "Spring 2026", SectionCode: &code, InstructorName: &instr},
			Course:   courses.EnrollmentCourse{ID: courseID, Department: "CPTS", Number: "322", Title: "SE I"},
			School:   courses.EnrollmentSchool{ID: schoolID, Acronym: "WSU"},
			Role:     courses.MemberRoleStudent,
			JoinedAt: joinedAt,
		}}, nil)

	req := authedRequestMethod(t, http.MethodGet, "/me/courses", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListMyEnrollmentsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Enrollments, 1)
	e := resp.Enrollments[0]
	assert.Equal(t, sectionID, uuid.UUID(e.Section.Id))
	assert.Equal(t, "Spring 2026", e.Section.Term)
	require.NotNil(t, e.Section.SectionCode)
	assert.Equal(t, "01", *e.Section.SectionCode)
	assert.Equal(t, courseID, uuid.UUID(e.Course.Id))
	assert.Equal(t, "CPTS", e.Course.Department)
	assert.Equal(t, schoolID, uuid.UUID(e.School.Id))
	assert.Equal(t, "WSU", e.School.Acronym)
	assert.Equal(t, api.EnrollmentResponseRoleStudent, e.Role)
	assert.True(t, e.JoinedAt.Equal(joinedAt))
}

func TestCoursesHandler_ListMyEnrollments_FiltersForwarded(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(p courses.ListMyEnrollmentsParams) bool {
			return p.Term != nil && *p.Term == "Spring 2026" &&
				p.Role != nil && *p.Role == courses.MemberRoleInstructor
		})).
		Return(nil, nil)

	req := authedRequestMethod(t, http.MethodGet, "/me/courses?term=Spring%202026&role=instructor", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Empty role= surfaces the same way as role=admin: oapi-codegen passes
// the empty string through, dbRoleFor("") returns false, service emits
// 400. Pinning this so a future reader knows the empty case is handled
// deliberately (and isn't accidentally treated as "no filter" the way
// term="" is).
func TestCoursesHandler_ListMyEnrollments_EmptyRoleRejected(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(p courses.ListMyEnrollmentsParams) bool {
			return p.Role != nil && *p.Role == courses.MemberRole("")
		})).
		Return(nil, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"role": "must be one of: student, instructor, ta",
		}))

	req := authedRequestMethod(t, http.MethodGet, "/me/courses?role=", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Per ASK-154 input-validation table, role=admin must surface as 400.
// oapi-codegen passes query-string enums through unvalidated (only path
// UUIDs get pre-handler validation), so the service is the layer that
// enforces the enum. The mock here returns the same AppError the real
// service would, asserting the handler propagates it as a 400 with the
// 'role: must be one of...' detail.
func TestCoursesHandler_ListMyEnrollments_BadRoleRejected(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(p courses.ListMyEnrollmentsParams) bool {
			return p.Role != nil && *p.Role == courses.MemberRole("admin")
		})).
		Return(nil, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
			"role": "must be one of: student, instructor, ta",
		}))

	req := authedRequestMethod(t, http.MethodGet, "/me/courses?role=admin", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "must be one of")
}

func TestCoursesHandler_ListMyEnrollments_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListMyEnrollments(mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	req := authedRequestMethod(t, http.MethodGet, "/me/courses", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ------------------------------------------------------------------------
// CheckMembership (ASK-148)
// ------------------------------------------------------------------------

func TestCoursesHandler_CheckMembership_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_CheckMembership_Enrolled(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	role := courses.MemberRoleStudent
	joinedAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		CheckMembership(mock.Anything, mock.Anything).
		Return(courses.MembershipCheck{Enrolled: true, Role: &role, JoinedAt: &joinedAt}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.MembershipCheckResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.True(t, resp.Enrolled)
	require.NotNil(t, resp.Role)
	assert.Equal(t, api.MembershipCheckResponseRoleStudent, *resp.Role)
	require.NotNil(t, resp.JoinedAt)
	assert.True(t, resp.JoinedAt.Equal(joinedAt))
}

// Critical contract test: not-enrolled must be 200 with explicit JSON
// nulls (not omitted). Verifies both the status code AND the wire
// shape so the frontend can safely destructure { enrolled, role,
// joined_at } without optional-chaining gymnastics.
func TestCoursesHandler_CheckMembership_NotEnrolledIs200WithNulls(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		CheckMembership(mock.Anything, mock.Anything).
		Return(courses.MembershipCheck{Enrolled: false}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Substring assertions pin the literal wire shape (proves the nulls
	// are emitted, not omitted -- "role": null vs missing key entirely).
	body := w.Body.String()
	assert.Contains(t, body, `"enrolled":false`)
	assert.Contains(t, body, `"role":null`)
	assert.Contains(t, body, `"joined_at":null`)
	// Typed decode + nil assertions guard against encoder drift -- if
	// the generated struct ever moved to a non-pointer field with
	// omitempty, the substring checks above would still pass on a body
	// that omits the keys, but this would catch it.
	var resp api.MembershipCheckResponse
	require.NoError(t, json.NewDecoder(strings.NewReader(body)).Decode(&resp))
	assert.False(t, resp.Enrolled)
	assert.Nil(t, resp.Role)
	assert.Nil(t, resp.JoinedAt)
}

func TestCoursesHandler_CheckMembership_CourseNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		CheckMembership(mock.Anything, mock.Anything).
		Return(courses.MembershipCheck{}, apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Course not found")
}

func TestCoursesHandler_CheckMembership_BadCourseUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/not-a-uuid/sections/%s/members/me", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoursesHandler_CheckMembership_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		CheckMembership(mock.Anything, mock.Anything).
		Return(courses.MembershipCheck{}, errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members/me", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ------------------------------------------------------------------------
// ListSectionMembers (ASK-143)
// ------------------------------------------------------------------------

func TestCoursesHandler_ListSectionMembers_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_ListSectionMembers_Empty(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return(courses.ListSectionMembersResult{}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListSectionMembersResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotNil(t, resp.Members)
	assert.Empty(t, resp.Members)
	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

func TestCoursesHandler_ListSectionMembers_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	userID := uuid.New()
	joinedAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return(courses.ListSectionMembersResult{
			Members: []courses.SectionMember{{
				UserID: userID, FirstName: "David", LastName: "Del Val",
				Role: courses.MemberRoleStudent, JoinedAt: joinedAt,
			}},
			HasMore: false,
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListSectionMembersResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Members, 1)
	m := resp.Members[0]
	assert.Equal(t, userID, uuid.UUID(m.UserId))
	assert.Equal(t, "David", m.FirstName)
	assert.Equal(t, "Del Val", m.LastName)
	assert.Equal(t, api.SectionMemberResponseRoleStudent, m.Role)
	assert.True(t, m.JoinedAt.Equal(joinedAt))
}

// Privacy floor regression: SectionMemberResponse must NEVER expose
// email, clerk_id, or any other user PII. The string-contains assertion
// pins the literal wire bytes -- if a future generated struct change
// surfaces an "email" field, this test fires immediately.
func TestCoursesHandler_ListSectionMembers_NoPIIInResponse(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return(courses.ListSectionMembersResult{
			Members: []courses.SectionMember{{
				UserID: uuid.New(), FirstName: "David", LastName: "Del Val",
				Role: courses.MemberRoleStudent, JoinedAt: time.Now().UTC(),
			}},
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.NotContains(t, body, `"email"`)
	assert.NotContains(t, body, `"clerk_id"`)
	assert.NotContains(t, body, `"clerkId"`)
}

func TestCoursesHandler_ListSectionMembers_FiltersForwarded(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(p courses.ListSectionMembersParams) bool {
			return p.Role != nil && *p.Role == courses.MemberRoleInstructor && p.Limit == 5
		})).
		Return(courses.ListSectionMembersResult{}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members?role=instructor&limit=5", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCoursesHandler_ListSectionMembers_BadCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections/%s/members?cursor=!!!notbase64", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "invalid cursor value")
	mockSvc.AssertNotCalled(t, "ListSectionMembers", mock.Anything, mock.Anything)
}

func TestCoursesHandler_ListSectionMembers_CursorRoundTrip(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	cursorJoinedAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	cursorUserID := uuid.New()
	token, err := courses.EncodeMemberCursor(courses.MemberCursor{JoinedAt: cursorJoinedAt, UserID: cursorUserID})
	require.NoError(t, err)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(p courses.ListSectionMembersParams) bool {
			return p.Cursor != nil && p.Cursor.UserID == cursorUserID && p.Cursor.JoinedAt.Equal(cursorJoinedAt)
		})).
		Return(courses.ListSectionMembersResult{}, nil)

	url := fmt.Sprintf("/courses/%s/sections/%s/members?cursor=%s", uuid.NewString(), uuid.NewString(), token)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCoursesHandler_ListSectionMembers_CourseNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return(courses.ListSectionMembersResult{}, apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Course not found")
}

func TestCoursesHandler_ListSectionMembers_BadCourseUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/not-a-uuid/sections/%s/members", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoursesHandler_ListSectionMembers_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return(courses.ListSectionMembersResult{}, errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/sections/%s/members", uuid.NewString(), uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================
// ListCourseSections tests (ASK-127)
// ============================================================

func TestCoursesHandler_ListCourseSections_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/sections", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoursesHandler_ListCourseSections_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodGet, "/courses/not-a-uuid/sections", nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCoursesHandler_ListCourseSections_TermTooLong_400 verifies
// that a term-length violation surfaces as 400 with details.term.
// oapi-codegen does NOT enforce maxLength on query strings (verified
// by tracing -- the wrapper passes the value through to the
// handler), so the validation lives in the service. This test
// pins the wire shape rather than the layer that enforces it.
func TestCoursesHandler_ListCourseSections_TermTooLong_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().ListCourseSections(mock.Anything, mock.Anything).
		Return(courses.ListCourseSectionsResult{}, apperrors.NewBadRequest(
			"Invalid query parameters",
			map[string]string{"term": "must be 30 characters or fewer"},
		))

	tooLong := strings.Repeat("a", 31)
	url := fmt.Sprintf("/courses/%s/sections?term=%s", uuid.NewString(), tooLong)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body api.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.NotNil(t, body.Details)
	assert.Contains(t, (*body.Details)["term"], "30")
}

func TestCoursesHandler_ListCourseSections_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().ListCourseSections(mock.Anything, mock.Anything).
		Return(courses.ListCourseSectionsResult{}, apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/sections", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	var body api.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "Course not found", body.Message)
}

func TestCoursesHandler_ListCourseSections_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().ListCourseSections(mock.Anything, mock.Anything).
		Return(courses.ListCourseSectionsResult{}, errors.New("connection refused"))

	url := fmt.Sprintf("/courses/%s/sections", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestCoursesHandler_ListCourseSections_Success_200 covers the
// happy path: service returns a populated result, handler renders
// 200 with the SectionResponse wire shape. Verifies the term
// query param is forwarded to the service in its decoded form.
func TestCoursesHandler_ListCourseSections_Success_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	courseID := uuid.New()
	sectionID := uuid.New()
	now := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	code := "01"
	instructor := "Dr. Ananth Jillepalli"

	mockSvc.EXPECT().ListCourseSections(mock.Anything,
		mock.MatchedBy(func(p courses.ListCourseSectionsParams) bool {
			return p.CourseID == courseID && p.Term != nil && *p.Term == "Spring 2026"
		})).Return(courses.ListCourseSectionsResult{
		Sections: []courses.SectionListing{
			{
				ID:             sectionID,
				CourseID:       courseID,
				Term:           "Spring 2026",
				SectionCode:    &code,
				InstructorName: &instructor,
				MemberCount:    34,
				CreatedAt:      now,
			},
		},
	}, nil)

	url := fmt.Sprintf("/courses/%s/sections?term=%s", courseID.String(), "Spring%202026")
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListCourseSectionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Sections, 1)
	assert.Equal(t, sectionID, uuid.UUID(resp.Sections[0].Id))
	assert.Equal(t, courseID, uuid.UUID(resp.Sections[0].CourseId))
	assert.Equal(t, "Spring 2026", resp.Sections[0].Term)
	require.NotNil(t, resp.Sections[0].SectionCode)
	assert.Equal(t, "01", *resp.Sections[0].SectionCode)
	require.NotNil(t, resp.Sections[0].InstructorName)
	assert.Equal(t, "Dr. Ananth Jillepalli", *resp.Sections[0].InstructorName)
	assert.Equal(t, int64(34), resp.Sections[0].MemberCount)
	// Required field per the openapi schema -- pin it in the
	// success wire test so a regression that drops it from the
	// mapper would be caught. coderabbit PR #160 feedback.
	assert.True(t, resp.Sections[0].CreatedAt.Equal(now))
}

// TestCoursesHandler_ListCourseSections_EmptyRendersBracket
// verifies the empty-result wire shape is `sections: []` (NOT
// null). Decode into a generic map so we catch a missing-key bug
// (the typed decode would happily accept a JSON null as a nil
// slice).
func TestCoursesHandler_ListCourseSections_EmptyRendersBracket(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().ListCourseSections(mock.Anything, mock.Anything).
		Return(courses.ListCourseSectionsResult{Sections: []courses.SectionListing{}}, nil)

	url := fmt.Sprintf("/courses/%s/sections", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	sectionsAny, ok := raw["sections"].([]any)
	require.True(t, ok, "sections must serialize as JSON array, not null")
	assert.Empty(t, sectionsAny)
}

// TestCoursesHandler_ListCourseSections_NullableFields verifies a
// section with NULL section_code + NULL instructor_name renders as
// JSON null on the wire (not missing keys, not zero-value empty
// strings).
func TestCoursesHandler_ListCourseSections_NullableFields(t *testing.T) {
	mockSvc := mock_handlers.NewMockCourseService(t)
	h := handlers.NewCoursesHandler(mockSvc)

	mockSvc.EXPECT().ListCourseSections(mock.Anything, mock.Anything).
		Return(courses.ListCourseSectionsResult{
			Sections: []courses.SectionListing{
				{
					ID:          uuid.New(),
					CourseID:    uuid.New(),
					Term:        "Spring 2026",
					MemberCount: 0,
					CreatedAt:   time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					// SectionCode + InstructorName left nil
				},
			},
		}, nil)

	url := fmt.Sprintf("/courses/%s/sections", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := coursesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	sectionsAny, ok := raw["sections"].([]any)
	require.True(t, ok)
	require.Len(t, sectionsAny, 1)
	first := sectionsAny[0].(map[string]any)
	assert.Nil(t, first["section_code"], "null section_code must render as JSON null")
	assert.Nil(t, first["instructor_name"], "null instructor_name must render as JSON null")
}
