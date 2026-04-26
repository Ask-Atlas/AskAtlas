package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/aiedits"
	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	gosse "github.com/tmaxmax/go-sse"
)

// editSystemPrompt is the system message sent on every AI edit
// request. Constant so OpenAI's auto-cache reuses the read across
// concurrent edits (>= 1024 tokens once we extend with retrieved
// chunks in ASK-223).
//
// Keeps the model out of "explain what you changed" territory --
// frontend computes diffs from `original_span` vs `replacement`
// (per ticket Decision: server returns full replacement, NOT a
// diff syntax).
const editSystemPrompt = `You are an inline editor for a student's study guide.

Rewrite ONLY the selected text per the instruction. Preserve the original markdown formatting, bullet structure, code blocks, and inline LaTeX/math. Do NOT add prefatory text such as "Here is the rewrite:" or "Sure, here you go". Do NOT explain your changes. Return ONLY the replacement text -- whatever you output will be substituted in-place where the selection used to be.`

// AIEditService is the slice of aiedits.Service the handler uses.
type AIEditService interface {
	RecordEdit(ctx context.Context, params aiedits.RecordEditParams) (aiedits.Edit, error)
	UpdateAcceptance(ctx context.Context, params aiedits.UpdateAcceptanceParams) (aiedits.Edit, error)
}

// AIEditStudyGuideReader is the slice of studyguides.Service used
// for the edit-permission check (creator-only at MVP; expand to
// grant-based when ASK-211 grows an `edit` permission).
type AIEditStudyGuideReader interface {
	GetStudyGuide(ctx context.Context, params studyguides.GetStudyGuideParams) (studyguides.StudyGuideDetail, error)
}

// AIEditStreamer is the slice of ai.Client used to dispatch the
// stream.
type AIEditStreamer interface {
	Stream(ctx context.Context, req ai.StreamRequest) (<-chan ai.Event, error)
}

// AIEditHandler serves POST /api/study-guides/{id}/ai/edit and
// PATCH /api/study-guides/{id}/ai/edits/{edit_id} (ASK-215).
type AIEditHandler struct {
	studyGuides AIEditStudyGuideReader
	edits       AIEditService
	streamer    AIEditStreamer
}

func NewAIEditHandler(reader AIEditStudyGuideReader, edits AIEditService, streamer AIEditStreamer) *AIEditHandler {
	return &AIEditHandler{
		studyGuides: reader,
		edits:       edits,
		streamer:    streamer,
	}
}

// AIEdit handles POST /api/study-guides/{study_guide_id}/ai/edit.
// Streams the model's replacement and persists an audit row on
// completion.
func (h *AIEditHandler) AIEdit(w http.ResponseWriter, r *http.Request, studyGuideID openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	guideID := uuid.UUID(studyGuideID)

	// Edit-permission gate. GetStudyGuide returns 404 if the viewer
	// can't see the guide; we additionally require creator ownership
	// to edit. Grants-based edit is a future expansion.
	detail, err := h.studyGuides.GetStudyGuide(r.Context(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID,
		ViewerID:     viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("AIEdit: GetStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}
	if detail.Creator.ID != viewerID {
		apperrors.RespondWithError(w, apperrors.NewForbidden())
		return
	}

	var body api.AIEditJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}
	if vErr := validateAIEditBody(body); vErr != nil {
		apperrors.RespondWithError(w, vErr)
		return
	}

	prompt := buildAIEditPrompt(body)
	model := ai.ModelDefault

	events, err := h.streamer.Stream(r.Context(), ai.StreamRequest{
		UserID:    viewerID,
		Feature:   ai.FeatureEdit,
		Model:     model,
		MaxTokens: 2048,
		System: []ai.Block{
			{Text: editSystemPrompt, CacheControl: true},
		},
		Messages: []ai.Message{
			{Role: ai.RoleUser, Blocks: []ai.Block{{Text: prompt}}},
		},
	})
	if err != nil {
		slog.Error("AIEdit: ai.Stream init failed", "error", err)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	// Custom stream loop: forwards each event to SSE AND captures
	// the replacement + final usage for the audit row. Can't reuse
	// ai.WriteStream because we need to inspect every event.
	session, sseErr := gosse.Upgrade(w, r)
	if sseErr != nil {
		slog.Error("AIEdit: sse upgrade failed", "error", sseErr)
		return
	}

	var replacement strings.Builder
	var captured ai.Usage

	heartbeat := time.NewTicker(ai.HeartbeatInterval)
	defer heartbeat.Stop()

streamLoop:
	for {
		select {
		case <-r.Context().Done():
			break streamLoop

		case <-heartbeat.C:
			if sendErr := session.Send(&gosse.Message{}); sendErr != nil {
				break streamLoop
			}
			_ = session.Flush()

		case ev, ok := <-events:
			if !ok {
				break streamLoop
			}
			switch ev.Kind {
			case ai.EventDelta:
				replacement.WriteString(ev.Delta)
				if writeAIEditEvent(session, ev) != nil {
					break streamLoop
				}
			case ai.EventUsage:
				if ev.Usage != nil {
					captured = *ev.Usage
				}
				if writeAIEditEvent(session, ev) != nil {
					break streamLoop
				}
			case ai.EventError:
				if writeAIEditEvent(session, ev) != nil {
					break streamLoop
				}
			case ai.EventDone:
				// Persist the audit row INLINE before emitting `done`
				// so we can include the edit_id in the payload. The
				// ASK-217 diff overlay needs this id to PATCH the user's
				// accept/reject outcome -- without it, the eval signal
				// for that user is lost. Detached context inside
				// persistAuditRow keeps the write robust to a client
				// tab closing in the ~50ms persist window.
				editID := h.persistAuditRow(guideID, viewerID, body, replacement.String(), model, captured)
				if writeAIEditDone(session, editID) != nil {
					break streamLoop
				}
			}
			heartbeat.Reset(ai.HeartbeatInterval)
		}
	}

	// On streams that didn't reach EventDone (mid-stream client
	// disconnect or upstream error) we deliberately do NOT persist:
	// the EventDone branch above is the ONLY place that calls
	// persistAuditRow, so recording a half-baked replacement the
	// user never saw is structurally impossible.
}

// persistAuditRow writes the study_guide_edits row for a successful
// AI edit stream. Returns the new row's ID (or uuid.Nil on failure)
// so the caller can include it in the `done` SSE event.
func (h *AIEditHandler) persistAuditRow(
	guideID uuid.UUID,
	viewerID uuid.UUID,
	body api.AIEditJSONRequestBody,
	replacement string,
	model ai.Model,
	usage ai.Usage,
) uuid.UUID {
	if replacement == "" {
		return uuid.Nil
	}
	insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	edit, err := h.edits.RecordEdit(insertCtx, aiedits.RecordEditParams{
		StudyGuideID:   guideID,
		UserID:         viewerID,
		Instruction:    body.Instruction,
		SelectionStart: int32(body.SelectionStart),
		SelectionEnd:   int32(body.SelectionEnd),
		OriginalSpan:   body.SelectionText,
		Replacement:    replacement,
		Model:          string(model),
		InputTokens:    usage.InputTokens,
		OutputTokens:   usage.OutputTokens,
	})
	if err != nil {
		slog.Error("AIEdit: persist audit row failed", "error", err, "study_guide_id", guideID, "user_id", viewerID)
		return uuid.Nil
	}
	return edit.ID
}

// UpdateAIEdit handles PATCH /api/study-guides/{study_guide_id}/ai/edits/{edit_id}.
func (h *AIEditHandler) UpdateAIEdit(
	w http.ResponseWriter,
	r *http.Request,
	studyGuideID openapi_types.UUID,
	editID openapi_types.UUID,
) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.UpdateAIEditJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	edit, err := h.edits.UpdateAcceptance(r.Context(), aiedits.UpdateAcceptanceParams{
		ID:           uuid.UUID(editID),
		StudyGuideID: uuid.UUID(studyGuideID),
		UserID:       viewerID,
		Accepted:     body.Accepted,
	})
	if err != nil {
		if errors.Is(err, aiedits.ErrNotFound) {
			apperrors.RespondWithError(w, apperrors.NewNotFound("Edit not found"))
			return
		}
		slog.Error("UpdateAIEdit failed", "error", err)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	respondJSON(w, http.StatusOK, mapEditToWire(edit))
}

func validateAIEditBody(body api.AIEditJSONRequestBody) *apperrors.AppError {
	if strings.TrimSpace(body.Instruction) == "" {
		return apperrors.NewBadRequest("instruction is required", nil)
	}
	if strings.TrimSpace(body.SelectionText) == "" {
		return apperrors.NewBadRequest("selection_text is required", nil)
	}
	if body.SelectionStart < 0 || body.SelectionEnd < body.SelectionStart {
		return apperrors.NewBadRequest("invalid selection range", nil)
	}
	return nil
}

// buildAIEditPrompt assembles the user message. Stable structure so
// OpenAI's auto-cache can reuse the system + leading preamble across
// consecutive edits on the same guide section.
func buildAIEditPrompt(body api.AIEditJSONRequestBody) string {
	var sb strings.Builder
	if body.DocContext != nil {
		if body.DocContext.Title != nil {
			sb.WriteString("Study guide: ")
			sb.WriteString(*body.DocContext.Title)
			sb.WriteString("\n\n")
		}
		if body.DocContext.Preceding != nil && *body.DocContext.Preceding != "" {
			sb.WriteString("Text immediately before the selection:\n")
			sb.WriteString(*body.DocContext.Preceding)
			sb.WriteString("\n\n")
		}
	}
	sb.WriteString("Selected text (rewrite this):\n")
	sb.WriteString(body.SelectionText)
	sb.WriteString("\n\n")
	if body.DocContext != nil && body.DocContext.Following != nil && *body.DocContext.Following != "" {
		sb.WriteString("Text immediately after the selection:\n")
		sb.WriteString(*body.DocContext.Following)
		sb.WriteString("\n\n")
	}
	sb.WriteString("Instruction: ")
	sb.WriteString(body.Instruction)
	return sb.String()
}

func writeAIEditEvent(session *gosse.Session, ev ai.Event) error {
	msg := &gosse.Message{Type: gosse.Type(string(ev.Kind))}
	payload, err := aiEditPayloadJSON(ev)
	if err != nil {
		return fmt.Errorf("encode %s payload: %w", ev.Kind, err)
	}
	msg.AppendData(string(payload))
	if err := session.Send(msg); err != nil {
		return err
	}
	return session.Flush()
}

// writeAIEditDone emits the terminal `done` event. The payload
// includes the persisted audit row's id so the frontend can PATCH
// the accept/reject outcome (ASK-217). `editID == uuid.Nil` means
// the persist failed -- we still emit `done` so the user sees the
// final reply, but `edit_id` is empty, which the frontend uses to
// hide the accept/reject controls.
func writeAIEditDone(session *gosse.Session, editID uuid.UUID) error {
	payload := struct {
		EditID string `json:"edit_id,omitempty"`
	}{}
	if editID != uuid.Nil {
		payload.EditID = editID.String()
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode done payload: %w", err)
	}
	msg := &gosse.Message{Type: gosse.Type(string(ai.EventDone))}
	msg.AppendData(string(encoded))
	if err := session.Send(msg); err != nil {
		return err
	}
	return session.Flush()
}

func aiEditPayloadJSON(ev ai.Event) ([]byte, error) {
	switch ev.Kind {
	case ai.EventDelta:
		return json.Marshal(struct {
			Text string `json:"text"`
		}{Text: ev.Delta})
	case ai.EventUsage:
		if ev.Usage == nil {
			return json.Marshal(ai.Usage{})
		}
		return json.Marshal(*ev.Usage)
	case ai.EventError:
		msg := "stream error"
		if ev.Err != nil {
			msg = ev.Err.Error()
		}
		return json.Marshal(struct {
			Message string `json:"message"`
		}{Message: msg})
	case ai.EventDone:
		return []byte("{}"), nil
	default:
		return nil, fmt.Errorf("unknown event kind %q", ev.Kind)
	}
}

func mapEditToWire(e aiedits.Edit) api.AIEditAuditRow {
	out := api.AIEditAuditRow{
		Id:             openapi_types.UUID(e.ID),
		StudyGuideId:   openapi_types.UUID(e.StudyGuideID),
		Instruction:    e.Instruction,
		SelectionStart: int(e.SelectionStart),
		SelectionEnd:   int(e.SelectionEnd),
		OriginalSpan:   e.OriginalSpan,
		Replacement:    e.Replacement,
		Model:          e.Model,
		InputTokens:    int(e.InputTokens),
		OutputTokens:   int(e.OutputTokens),
		CreatedAt:      e.CreatedAt,
	}
	if e.Accepted != nil {
		v := *e.Accepted
		out.Accepted = &v
	}
	if e.AcceptedAt != nil {
		t := *e.AcceptedAt
		out.AcceptedAt = &t
	}
	return out
}
