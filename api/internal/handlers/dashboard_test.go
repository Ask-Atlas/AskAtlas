package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/dashboard"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// dashboardTestRouter wires the composite handler with mocked
// sibling services so /me/dashboard requests resolve through the
// same routing the real binary uses. The DashboardHandler under
// test is the only real (non-mock) handler.
func dashboardTestRouter(t *testing.T, dh *handlers.DashboardHandler) chi.Router {
	fileH := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fileH, gh, sh, ch, sgh, nil, qh, ssh, nil, nil, dh, nil, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestDashboardHandler_ListDashboard_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockDashboardService(t)
	h := handlers.NewDashboardHandler(mockSvc)

	// No authctx -> handler must short-circuit with 401 before
	// touching the service mock.
	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	w := httptest.NewRecorder()
	r := dashboardTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDashboardHandler_ListDashboard_Success_FullPayload(t *testing.T) {
	mockSvc := mock_handlers.NewMockDashboardService(t)
	h := handlers.NewDashboardHandler(mockSvc)

	viewer := uuid.New()
	courseID := uuid.New()
	guideID := uuid.New()
	sessionID := uuid.New()
	fileID := uuid.New()
	term := "Spring 2026"
	updatedAt := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		GetDashboard(mock.Anything, mock.MatchedBy(func(p dashboard.GetDashboardParams) bool {
			return p.ViewerID == viewer
		})).
		Return(dashboard.DashboardData{
			Courses: dashboard.DashboardCoursesSection{
				EnrolledCount: 1,
				CurrentTerm:   &term,
				Courses: []dashboard.DashboardCourseSummary{{
					ID:          courseID,
					Department:  "CPTS",
					Number:      "322",
					Title:       "SE I",
					Role:        dashboard.MemberRoleStudent,
					SectionTerm: term,
				}},
			},
			StudyGuides: dashboard.DashboardStudyGuidesSection{
				CreatedCount: 3,
				Recent: []dashboard.DashboardStudyGuideSummary{{
					ID:               guideID,
					Title:            "Binary Trees",
					CourseDepartment: "CPTS",
					CourseNumber:     "322",
					UpdatedAt:        updatedAt,
				}},
			},
			Practice: dashboard.DashboardPracticeSection{
				SessionsCompleted:      12,
				TotalQuestionsAnswered: 87,
				OverallAccuracy:        74,
				RecentSessions: []dashboard.DashboardSessionSummary{{
					ID:              sessionID,
					QuizTitle:       "Tree Traversal Quiz",
					StudyGuideTitle: "Binary Trees",
					ScorePercentage: 80,
					CompletedAt:     completedAt,
				}},
			},
			Files: dashboard.DashboardFilesSection{
				TotalCount: 15,
				TotalSize:  52428800,
				Recent: []dashboard.DashboardFileSummary{{
					ID:        fileID,
					Name:      "Lecture Notes Week 12.pdf",
					MimeType:  "application/pdf",
					UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
				}},
			},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := dashboardTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.DashboardResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	// Courses section.
	assert.Equal(t, int32(1), resp.Courses.EnrolledCount)
	require.NotNil(t, resp.Courses.CurrentTerm)
	assert.Equal(t, term, *resp.Courses.CurrentTerm)
	require.Len(t, resp.Courses.Courses, 1)
	assert.Equal(t, courseID, resp.Courses.Courses[0].Id)
	assert.Equal(t, api.DashboardCourseSummaryRole("student"), resp.Courses.Courses[0].Role)

	// Study guides section.
	assert.Equal(t, int32(3), resp.StudyGuides.CreatedCount)
	require.Len(t, resp.StudyGuides.Recent, 1)
	assert.Equal(t, "Binary Trees", resp.StudyGuides.Recent[0].Title)

	// Practice section.
	assert.Equal(t, int32(12), resp.Practice.SessionsCompleted)
	assert.Equal(t, int32(87), resp.Practice.TotalQuestionsAnswered)
	assert.Equal(t, int32(74), resp.Practice.OverallAccuracy)
	require.Len(t, resp.Practice.RecentSessions, 1)
	assert.Equal(t, int32(80), resp.Practice.RecentSessions[0].ScorePercentage)

	// Files section.
	assert.Equal(t, int32(15), resp.Files.TotalCount)
	assert.Equal(t, int64(52428800), resp.Files.TotalSize)
	require.Len(t, resp.Files.Recent, 1)
	assert.Equal(t, "application/pdf", resp.Files.Recent[0].MimeType)
}

func TestDashboardHandler_ListDashboard_EmptyData_RendersFullEnvelope(t *testing.T) {
	mockSvc := mock_handlers.NewMockDashboardService(t)
	h := handlers.NewDashboardHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().GetDashboard(mock.Anything, mock.Anything).Return(dashboard.DashboardData{
		Courses: dashboard.DashboardCoursesSection{
			Courses: []dashboard.DashboardCourseSummary{},
		},
		StudyGuides: dashboard.DashboardStudyGuidesSection{
			Recent: []dashboard.DashboardStudyGuideSummary{},
		},
		Practice: dashboard.DashboardPracticeSection{
			RecentSessions: []dashboard.DashboardSessionSummary{},
		},
		Files: dashboard.DashboardFilesSection{
			Recent: []dashboard.DashboardFileSummary{},
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := dashboardTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Wire contract: every list field must render as [], not null.
	// Pinned by raw-substring assertions because typed unmarshal
	// can't distinguish nil-slice from empty-slice from null.
	assert.Contains(t, body, `"courses":[]`)
	assert.Contains(t, body, `"recent":[]`) // appears 2x (study_guides.recent, files.recent)
	assert.Contains(t, body, `"recent_sessions":[]`)

	// current_term is nullable; when nil it must render as null
	// rather than be absent (the schema declares it required+nullable).
	assert.Contains(t, body, `"current_term":null`)
}

func TestDashboardHandler_ListDashboard_ServiceFails_Returns500(t *testing.T) {
	mockSvc := mock_handlers.NewMockDashboardService(t)
	h := handlers.NewDashboardHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().
		GetDashboard(mock.Anything, mock.Anything).
		Return(dashboard.DashboardData{}, errors.New("db is down"))

	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := dashboardTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
