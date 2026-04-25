package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/refs"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func refsTestRouter(t *testing.T, rh *handlers.RefsHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, nil, qh, ssh, nil, nil, nil, rh, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestRefsHandler_ResolveRefs_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockRefsService(t)
	h := handlers.NewRefsHandler(mockSvc)

	viewerID := uuid.New()
	sgID := uuid.New()
	fileID := uuid.New()
	missingCourseID := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	qc := 2
	rec := true
	fsize := int64(2048)

	mockSvc.EXPECT().
		Resolve(mock.Anything, viewerID, mock.MatchedBy(func(in []refs.Ref) bool {
			return len(in) == 3
		})).
		Return(map[string]*refs.Summary{
			refs.Key(refs.TypeStudyGuide, sgID): {
				Type:          refs.TypeStudyGuide,
				ID:            sgID,
				Title:         "BST primer",
				Course:        &refs.CourseInfo{Department: "CPTS", Number: "322"},
				QuizCount:     &qc,
				IsRecommended: &rec,
			},
			refs.Key(refs.TypeFile, fileID): {
				Type:     refs.TypeFile,
				ID:       fileID,
				Name:     "notes.pdf",
				Size:     &fsize,
				MimeType: "application/pdf",
				Status:   "complete",
			},
			// Third ref (unknown course) surfaces as nil -> JSON null.
			refs.Key(refs.TypeCourse, missingCourseID): nil,
		}, nil)

	body := fmt.Sprintf(`{"refs":[
		{"type":"sg","id":"%s"},
		{"type":"file","id":"%s"},
		{"type":"course","id":"%s"}
	]}`, sgID, fileID, missingCourseID)

	req := httptest.NewRequest(http.MethodPost, "/refs/resolve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), viewerID))
	w := httptest.NewRecorder()
	refsTestRouter(t, h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Results map[string]*json.RawMessage `json:"results"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp.Results, 3)

	// Unresolved ref must encode as JSON null.
	nilKey := refs.Key(refs.TypeCourse, missingCourseID)
	nilRaw, ok := resp.Results[nilKey]
	require.True(t, ok, "nil entry must be present under its key")
	require.True(t,
		nilRaw == nil || bytes.Equal(*nilRaw, []byte("null")),
		"nil ref must encode as JSON null",
	)

	// Populated entry decodes back to a RefSummary with the expected fields.
	sgRaw := resp.Results[refs.Key(refs.TypeStudyGuide, sgID)]
	require.NotNil(t, sgRaw)
	var sg api.RefSummary
	require.NoError(t, json.Unmarshal(*sgRaw, &sg))
	assert.Equal(t, api.RefSummaryTypeSg, sg.Type)
	require.NotNil(t, sg.Title)
	assert.Equal(t, "BST primer", *sg.Title)
	require.NotNil(t, sg.QuizCount)
	assert.Equal(t, 2, *sg.QuizCount)
	require.NotNil(t, sg.IsRecommended)
	assert.True(t, *sg.IsRecommended)
}

func TestRefsHandler_ResolveRefs_Unauthorized_401(t *testing.T) {
	mockSvc := mock_handlers.NewMockRefsService(t)
	h := handlers.NewRefsHandler(mockSvc)

	body := `{"refs":[{"type":"sg","id":"00000000-0000-0000-0000-000000000000"}]}`
	req := httptest.NewRequest(http.MethodPost, "/refs/resolve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// no authctx on purpose
	w := httptest.NewRecorder()
	refsTestRouter(t, h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Body-schema validation (maxItems 50, type enum) lives in the OAPI
// validator middleware wired one layer up in cmd/api/main.go, not in
// the handler -- the handler trusts its bindings. e2e covers those
// paths end-to-end against the deployed validator.

func TestRefsHandler_ResolveRefs_MalformedBody_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockRefsService(t)
	h := handlers.NewRefsHandler(mockSvc)

	body := `{not json`
	req := httptest.NewRequest(http.MethodPost, "/refs/resolve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()
	refsTestRouter(t, h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefsHandler_ResolveRefs_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockRefsService(t)
	h := handlers.NewRefsHandler(mockSvc)

	mockSvc.EXPECT().
		Resolve(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db exploded"))

	body := fmt.Sprintf(`{"refs":[{"type":"sg","id":"%s"}]}`, uuid.New())
	req := httptest.NewRequest(http.MethodPost, "/refs/resolve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()
	refsTestRouter(t, h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
