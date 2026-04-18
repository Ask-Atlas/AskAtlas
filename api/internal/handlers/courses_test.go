package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
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
