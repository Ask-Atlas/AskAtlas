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
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/schools"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// schoolsTestRouter wires the composite handler with mocked file/grant services
// so /schools requests resolve through the same routing the real binary uses.
func schoolsTestRouter(t *testing.T, sh *handlers.SchoolsHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func authedRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	ctx := authctx.WithUserID(req.Context(), uuid.New())
	return req.WithContext(ctx)
}

func TestSchoolsHandler_ListSchools_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	// No authctx.WithUserID -> handler should reject before touching the service.
	req := httptest.NewRequest(http.MethodGet, "/schools", nil)
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSchoolsHandler_ListSchools_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	schoolID := uuid.New()
	domain := "wsu.edu"
	city := "Pullman"
	state := "WA"
	country := "US"
	created := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListSchools(mock.Anything, mock.Anything).
		Return(schools.ListSchoolsResult{
			Schools: []schools.School{{
				ID:        schoolID,
				Name:      "Washington State University",
				Acronym:   "WSU",
				Domain:    &domain,
				City:      &city,
				State:     &state,
				Country:   &country,
				CreatedAt: created,
			}},
			HasMore:    false,
			NextCursor: nil,
		}, nil)

	req := authedRequest(t, "/schools")
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListSchoolsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	require.Len(t, resp.Schools, 1)
	assert.Equal(t, schoolID, uuid.UUID(resp.Schools[0].Id))
	assert.Equal(t, "Washington State University", resp.Schools[0].Name)
	assert.Equal(t, "WSU", resp.Schools[0].Acronym)
	require.NotNil(t, resp.Schools[0].Domain)
	assert.Equal(t, "wsu.edu", *resp.Schools[0].Domain)
	require.NotNil(t, resp.Schools[0].Country)
	assert.Equal(t, "US", *resp.Schools[0].Country)
	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

func TestSchoolsHandler_ListSchools_Empty(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	mockSvc.EXPECT().
		ListSchools(mock.Anything, mock.Anything).
		Return(schools.ListSchoolsResult{Schools: nil, HasMore: false, NextCursor: nil}, nil)

	req := authedRequest(t, "/schools?q=xyznonexistent")
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListSchoolsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Empty(t, resp.Schools)
	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

func TestSchoolsHandler_ListSchools_BadCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	// Service must not be called when the cursor decode fails.
	req := authedRequest(t, "/schools?cursor=!!!notbase64")
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	details, ok := body["details"].(map[string]any)
	require.True(t, ok, "expected details object in error response")
	assert.Equal(t, "invalid cursor value", details["cursor"])

	// Cursor decode failure must short-circuit before the service is invoked.
	mockSvc.AssertNotCalled(t, "ListSchools", mock.Anything, mock.Anything)
}

func TestSchoolsHandler_ListSchools_CursorForwardedToService(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	cursorID := uuid.New()
	cursor := schools.Cursor{Name: "Stanford University", ID: cursorID}
	token, err := schools.EncodeCursor(cursor)
	require.NoError(t, err)

	mockSvc.EXPECT().
		ListSchools(mock.Anything, mock.MatchedBy(func(p schools.ListSchoolsParams) bool {
			return p.Cursor != nil && p.Cursor.Name == "Stanford University" && p.Cursor.ID == cursorID &&
				p.Q != nil && *p.Q == "wsu" &&
				p.Limit == 10
		})).
		Return(schools.ListSchoolsResult{}, nil)

	req := authedRequest(t, "/schools?q=wsu&page_limit=10&cursor="+token)
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSchoolsHandler_GetSchool_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	// No authctx -> handler should reject before touching the service.
	req := httptest.NewRequest(http.MethodGet, "/schools/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSchoolsHandler_GetSchool_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	schoolID := uuid.New()
	domain := "wsu.edu"
	city := "Pullman"
	state := "WA"
	country := "US"
	created := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		GetSchool(mock.Anything, mock.MatchedBy(func(p schools.GetSchoolParams) bool {
			return p.SchoolID == schoolID
		})).
		Return(schools.School{
			ID:        schoolID,
			Name:      "Washington State University",
			Acronym:   "WSU",
			Domain:    &domain,
			City:      &city,
			State:     &state,
			Country:   &country,
			CreatedAt: created,
		}, nil)

	req := authedRequest(t, "/schools/"+schoolID.String())
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.SchoolResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, schoolID, uuid.UUID(resp.Id))
	assert.Equal(t, "Washington State University", resp.Name)
	assert.Equal(t, "WSU", resp.Acronym)
	require.NotNil(t, resp.Domain)
	assert.Equal(t, "wsu.edu", *resp.Domain)
}

func TestSchoolsHandler_GetSchool_NotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	mockSvc.EXPECT().
		GetSchool(mock.Anything, mock.Anything).
		Return(schools.School{}, fmt.Errorf("GetSchool: %w", apperrors.ErrNotFound))

	req := authedRequest(t, "/schools/"+uuid.New().String())
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	// Spec mandates the school-specific message rather than the generic "Resource not found".
	assert.Equal(t, "School not found", body["message"])
}

func TestSchoolsHandler_GetSchool_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	mockSvc.EXPECT().
		GetSchool(mock.Anything, mock.Anything).
		Return(schools.School{}, errors.New("db down"))

	req := authedRequest(t, "/schools/"+uuid.New().String())
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSchoolsHandler_ListSchools_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockSchoolService(t)
	h := handlers.NewSchoolsHandler(mockSvc)

	mockSvc.EXPECT().
		ListSchools(mock.Anything, mock.Anything).
		Return(schools.ListSchoolsResult{}, errors.New("db down"))

	req := authedRequest(t, "/schools")
	w := httptest.NewRecorder()

	r := schoolsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
