package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/aiedits"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type fakeStudyGuideReader struct {
	creatorID uuid.UUID
	err       error
}

func (f *fakeStudyGuideReader) GetStudyGuide(_ context.Context, _ studyguides.GetStudyGuideParams) (studyguides.StudyGuideDetail, error) {
	if f.err != nil {
		return studyguides.StudyGuideDetail{}, f.err
	}
	return studyguides.StudyGuideDetail{
		Creator: studyguides.Creator{ID: f.creatorID},
	}, nil
}

type fakeAIEditService struct {
	recordedParams aiedits.RecordEditParams
	recordCalls    int
	recordErr      error
	recordReturnID uuid.UUID
	updateErr      error
	lastUpdate     aiedits.UpdateAcceptanceParams
}

func (f *fakeAIEditService) RecordEdit(_ context.Context, params aiedits.RecordEditParams) (aiedits.Edit, error) {
	f.recordCalls++
	f.recordedParams = params
	if f.recordErr != nil {
		return aiedits.Edit{}, f.recordErr
	}
	id := f.recordReturnID
	if id == uuid.Nil {
		id = uuid.New()
	}
	return aiedits.Edit{ID: id, StudyGuideID: params.StudyGuideID, UserID: params.UserID}, nil
}

func (f *fakeAIEditService) UpdateAcceptance(_ context.Context, params aiedits.UpdateAcceptanceParams) (aiedits.Edit, error) {
	f.lastUpdate = params
	if f.updateErr != nil {
		return aiedits.Edit{}, f.updateErr
	}
	acc := params.Accepted
	return aiedits.Edit{
		ID:           params.ID,
		StudyGuideID: params.StudyGuideID,
		Accepted:     &acc,
	}, nil
}

type fakeStreamer struct {
	events []ai.Event
	err    error
	called bool
	gotReq ai.StreamRequest
}

func (f *fakeStreamer) Stream(_ context.Context, req ai.StreamRequest) (<-chan ai.Event, error) {
	f.called = true
	f.gotReq = req
	if f.err != nil {
		return nil, f.err
	}
	ch := make(chan ai.Event, len(f.events))
	for _, ev := range f.events {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

func aiEditBody(t *testing.T, instr, sel string) []byte {
	t.Helper()
	body := map[string]any{
		"selection_text":  sel,
		"selection_start": 100,
		"selection_end":   100 + len(sel),
		"instruction":     instr,
	}
	out, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAIEdit_Stream_PersistsAuditRow(t *testing.T) {
	t.Parallel()

	creator := uuid.New()
	guideID := uuid.New()
	editID := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	svc := &fakeAIEditService{recordReturnID: editID}
	streamer := &fakeStreamer{events: []ai.Event{
		{Kind: ai.EventDelta, Delta: "Hel"},
		{Kind: ai.EventDelta, Delta: "lo "},
		{Kind: ai.EventDelta, Delta: "world."},
		{Kind: ai.EventUsage, Usage: &ai.Usage{InputTokens: 50, OutputTokens: 12}},
		{Kind: ai.EventDone},
	}}
	h := handlers.NewAIEditHandler(reader, svc, streamer)

	body := aiEditBody(t, "make it shorter", "Once upon a time there was a thing.")
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/study-guides/%s/ai/edit", guideID), strings.NewReader(string(body))).
		WithContext(authctx.WithUserID(context.Background(), creator))
	rec := httptest.NewRecorder()

	h.AIEdit(rec, req, openapi_types.UUID(guideID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%q", rec.Code, rec.Body.String())
	}
	wireBody := rec.Body.String()
	wantEditID := fmt.Sprintf(`"edit_id":%q`, editID.String())
	for _, want := range []string{
		`"text":"Hel"`, `"text":"lo "`, `"text":"world."`,
		`"input_tokens":50`,
		"event: done",
		wantEditID,
	} {
		if !strings.Contains(wireBody, want) {
			t.Errorf("missing %q in body:\n%s", want, wireBody)
		}
	}

	if svc.recordCalls != 1 {
		t.Fatalf("RecordEdit calls = %d, want 1", svc.recordCalls)
	}
	got := svc.recordedParams
	if got.Replacement != "Hello world." {
		t.Errorf("Replacement = %q, want %q", got.Replacement, "Hello world.")
	}
	if got.InputTokens != 50 || got.OutputTokens != 12 {
		t.Errorf("token counts = (%d, %d), want (50, 12)", got.InputTokens, got.OutputTokens)
	}
	if got.StudyGuideID != guideID {
		t.Errorf("StudyGuideID = %v, want %v", got.StudyGuideID, guideID)
	}
	if got.UserID != creator {
		t.Errorf("UserID = %v, want %v", got.UserID, creator)
	}
}

func TestAIEdit_DoneOmitsEditID_WhenPersistFails(t *testing.T) {
	t.Parallel()

	creator := uuid.New()
	guideID := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	svc := &fakeAIEditService{recordErr: errors.New("db unreachable")}
	streamer := &fakeStreamer{events: []ai.Event{
		{Kind: ai.EventDelta, Delta: "ok"},
		{Kind: ai.EventDone},
	}}
	h := handlers.NewAIEditHandler(reader, svc, streamer)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/study-guides/%s/ai/edit", guideID), strings.NewReader(string(aiEditBody(t, "do x", "y")))).
		WithContext(authctx.WithUserID(context.Background(), creator))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(guideID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	wireBody := rec.Body.String()
	if !strings.Contains(wireBody, "event: done") {
		t.Fatalf("missing done event in body:\n%s", wireBody)
	}
	// `omitempty` on the EditID JSON tag means the field should not
	// appear at all when the persist failed.
	if strings.Contains(wireBody, `"edit_id"`) {
		t.Errorf("done payload included edit_id despite persist failure:\n%s", wireBody)
	}
}

func TestAIEdit_Forbidden_NonCreator(t *testing.T) {
	t.Parallel()

	creator := uuid.New()
	viewer := uuid.New()
	guideID := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	svc := &fakeAIEditService{}
	streamer := &fakeStreamer{}
	h := handlers.NewAIEditHandler(reader, svc, streamer)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(string(aiEditBody(t, "x", "y")))).
		WithContext(authctx.WithUserID(context.Background(), viewer))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(guideID))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if streamer.called {
		t.Error("streamer was called despite 403")
	}
	if svc.recordCalls != 0 {
		t.Errorf("RecordEdit called %d times, want 0", svc.recordCalls)
	}
}

func TestAIEdit_NotFound_PropagatesFromStudyGuides(t *testing.T) {
	t.Parallel()

	reader := &fakeStudyGuideReader{err: apperrors.NewNotFound("Study guide not found")}
	h := handlers.NewAIEditHandler(reader, &fakeAIEditService{}, &fakeStreamer{})

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(string(aiEditBody(t, "x", "y")))).
		WithContext(authctx.WithUserID(context.Background(), uuid.New()))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(uuid.New()))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestAIEdit_Unauthenticated(t *testing.T) {
	t.Parallel()

	h := handlers.NewAIEditHandler(&fakeStudyGuideReader{}, &fakeAIEditService{}, &fakeStreamer{})
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(string(aiEditBody(t, "x", "y"))))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(uuid.New()))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAIEdit_BadBody_400(t *testing.T) {
	t.Parallel()

	creator := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	h := handlers.NewAIEditHandler(reader, &fakeAIEditService{}, &fakeStreamer{})

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{not json`)).
		WithContext(authctx.WithUserID(context.Background(), creator))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(uuid.New()))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400, body=%q", rec.Code, rec.Body.String())
	}
}

func TestAIEdit_StreamDisconnectBeforeDone_NoAuditRow(t *testing.T) {
	t.Parallel()

	// Channel closes WITHOUT ever sending EventDone -- mimics a
	// client disconnect or upstream cutoff after partial deltas.
	// We received some text on the wire but no completion signal,
	// so the audit row would be misleading. Don't persist.
	creator := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	svc := &fakeAIEditService{}
	streamer := &fakeStreamer{events: []ai.Event{
		{Kind: ai.EventDelta, Delta: "partial..."},
		// no EventDone, no EventError -- channel just closes
	}}
	h := handlers.NewAIEditHandler(reader, svc, streamer)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(string(aiEditBody(t, "x", "y")))).
		WithContext(authctx.WithUserID(context.Background(), creator))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(uuid.New()))

	if svc.recordCalls != 0 {
		t.Errorf("RecordEdit called %d times after disconnect-before-done, want 0", svc.recordCalls)
	}
}

func TestAIEdit_StreamError_NoAuditRow(t *testing.T) {
	t.Parallel()

	creator := uuid.New()
	reader := &fakeStudyGuideReader{creatorID: creator}
	svc := &fakeAIEditService{}
	streamer := &fakeStreamer{events: []ai.Event{
		{Kind: ai.EventError, Err: errors.New("upstream model error")},
	}}
	h := handlers.NewAIEditHandler(reader, svc, streamer)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(string(aiEditBody(t, "x", "y")))).
		WithContext(authctx.WithUserID(context.Background(), creator))
	rec := httptest.NewRecorder()
	h.AIEdit(rec, req, openapi_types.UUID(uuid.New()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (SSE upgrade succeeded; error is in-stream)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "event: error") {
		t.Errorf("body missing error event:\n%s", rec.Body.String())
	}
	if svc.recordCalls != 0 {
		t.Errorf("RecordEdit called %d times after stream error, want 0", svc.recordCalls)
	}
}

func TestUpdateAIEdit_Success(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	guideID := uuid.New()
	editID := uuid.New()
	svc := &fakeAIEditService{}
	h := handlers.NewAIEditHandler(&fakeStudyGuideReader{}, svc, &fakeStreamer{})

	req := httptest.NewRequest(http.MethodPatch, "/x", strings.NewReader(`{"accepted":true}`)).
		WithContext(authctx.WithUserID(context.Background(), user))
	rec := httptest.NewRecorder()
	h.UpdateAIEdit(rec, req, openapi_types.UUID(guideID), openapi_types.UUID(editID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%q", rec.Code, rec.Body.String())
	}
	if svc.lastUpdate.ID != editID {
		t.Errorf("ID = %v, want %v", svc.lastUpdate.ID, editID)
	}
	if !svc.lastUpdate.Accepted {
		t.Error("Accepted = false, want true")
	}
	if svc.lastUpdate.UserID != user {
		t.Errorf("UserID = %v, want %v", svc.lastUpdate.UserID, user)
	}
}

func TestUpdateAIEdit_NotFound(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	svc := &fakeAIEditService{updateErr: aiedits.ErrNotFound}
	h := handlers.NewAIEditHandler(&fakeStudyGuideReader{}, svc, &fakeStreamer{})

	req := httptest.NewRequest(http.MethodPatch, "/x", strings.NewReader(`{"accepted":true}`)).
		WithContext(authctx.WithUserID(context.Background(), user))
	rec := httptest.NewRecorder()
	h.UpdateAIEdit(rec, req, openapi_types.UUID(uuid.New()), openapi_types.UUID(uuid.New()))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
