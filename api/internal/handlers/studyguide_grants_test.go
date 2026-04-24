package handlers_test

import (
	"bytes"
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
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// studyGuideGrantTestRouter wires the composite handler with a mocked
// StudyGuideGrantService so the chi route resolves through the same
// path the real binary uses.
func studyGuideGrantTestRouter(t *testing.T, h *handlers.StudyGuideGrantHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, h, qh, ssh, nil, nil, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func authedJSONRequest(t *testing.T, method, url string, body any, viewerID uuid.UUID) *http.Request {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, url, buf)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), viewerID))
	return req
}

func TestStudyGuideGrantHandler_Create_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	url := fmt.Sprintf("/study-guides/%s/grants", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStudyGuideGrantHandler_Create_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	viewerID := uuid.New()
	studyGuideID := uuid.New()
	granteeID := uuid.New()
	grantID := uuid.New()

	mockSvc.EXPECT().
		CreateGrant(mock.Anything, mock.MatchedBy(func(p studyguides.CreateGrantParams) bool {
			return p.StudyGuideID == studyGuideID &&
				p.ViewerID == viewerID &&
				p.GranteeType == "user" &&
				p.GranteeID == granteeID &&
				p.Permission == "view"
		})).
		Return(studyguides.Grant{
			ID:           grantID,
			StudyGuideID: studyGuideID,
			GranteeType:  "user",
			GranteeID:    granteeID,
			Permission:   "view",
			GrantedBy:    viewerID,
			CreatedAt:    time.Now().UTC(),
		}, nil)

	body := api.StudyGuideCreateGrantRequest{
		GranteeType: api.StudyGuideCreateGrantRequestGranteeTypeUser,
		GranteeId:   granteeID,
		Permission:  api.StudyGuideCreateGrantRequestPermissionView,
	}
	url := fmt.Sprintf("/study-guides/%s/grants", studyGuideID)
	req := authedJSONRequest(t, http.MethodPost, url, body, viewerID)
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestStudyGuideGrantHandler_Create_Forbidden(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	mockSvc.EXPECT().
		CreateGrant(mock.Anything, mock.Anything).
		Return(studyguides.Grant{}, apperrors.NewForbidden())

	body := api.StudyGuideCreateGrantRequest{
		GranteeType: api.StudyGuideCreateGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.StudyGuideCreateGrantRequestPermissionView,
	}
	url := fmt.Sprintf("/study-guides/%s/grants", uuid.New())
	req := authedJSONRequest(t, http.MethodPost, url, body, uuid.New())
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudyGuideGrantHandler_Revoke_NotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	mockSvc.EXPECT().
		RevokeGrant(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Grant not found"))

	body := api.StudyGuideRevokeGrantRequest{
		GranteeType: api.StudyGuideRevokeGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.StudyGuideRevokeGrantRequestPermissionView,
	}
	url := fmt.Sprintf("/study-guides/%s/grants", uuid.New())
	req := authedJSONRequest(t, http.MethodDelete, url, body, uuid.New())
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStudyGuideGrantHandler_Revoke_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	mockSvc.EXPECT().
		RevokeGrant(mock.Anything, mock.Anything).
		Return(nil)

	body := api.StudyGuideRevokeGrantRequest{
		GranteeType: api.StudyGuideRevokeGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.StudyGuideRevokeGrantRequestPermissionView,
	}
	url := fmt.Sprintf("/study-guides/%s/grants", uuid.New())
	req := authedJSONRequest(t, http.MethodDelete, url, body, uuid.New())
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestStudyGuideGrantHandler_List_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	studyGuideID := uuid.New()
	grantID := uuid.New()
	viewerID := uuid.New()

	mockSvc.EXPECT().
		ListGrants(mock.Anything, mock.Anything).
		Return([]studyguides.Grant{
			{
				ID:           grantID,
				StudyGuideID: studyGuideID,
				GranteeType:  "course",
				GranteeID:    uuid.New(),
				Permission:   "view",
				GrantedBy:    viewerID,
				CreatedAt:    time.Now().UTC(),
			},
		}, nil)

	url := fmt.Sprintf("/study-guides/%s/grants", studyGuideID)
	req := authedJSONRequest(t, http.MethodGet, url, nil, viewerID)
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListStudyGuideGrantsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Grants, 1)
	assert.Equal(t, "course", resp.Grants[0].GranteeType)
}

func TestStudyGuideGrantHandler_List_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideGrantService(t)
	h := handlers.NewStudyGuideGrantHandler(mockSvc)

	mockSvc.EXPECT().
		ListGrants(mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	url := fmt.Sprintf("/study-guides/%s/grants", uuid.New())
	req := authedJSONRequest(t, http.MethodGet, url, nil, uuid.New())
	w := httptest.NewRecorder()
	r := studyGuideGrantTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
