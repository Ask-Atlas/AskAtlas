package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/refs"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// RefsService is the application logic required by the RefsHandler.
type RefsService interface {
	Resolve(ctx context.Context, viewerID uuid.UUID, refs []refs.Ref) (map[string]*refs.Summary, error)
}

// RefsHandler manages the batch inline-ref resolution endpoint.
type RefsHandler struct {
	service RefsService
}

// NewRefsHandler creates a new RefsHandler backed by the given RefsService.
func NewRefsHandler(service RefsService) *RefsHandler {
	return &RefsHandler{service: service}
}

// ResolveRefs handles POST /api/refs/resolve (ASK-208).
func (h *RefsHandler) ResolveRefs(w http.ResponseWriter, r *http.Request) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.ResolveRefsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	domainRefs := make([]refs.Ref, 0, len(body.Refs))
	for _, ref := range body.Refs {
		domainRefs = append(domainRefs, refs.Ref{
			Type: refs.RefType(ref.Type),
			ID:   uuid.UUID(ref.Id),
		})
	}

	results, err := h.service.Resolve(r.Context(), viewerID, domainRefs)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ResolveRefs failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	// Use a *RefSummary map so nil entries serialize as JSON null,
	// matching the schema's `nullable: true` on additionalProperties.
	// The generated api.RefsResolveResponse.Results is a value map
	// which would emit zero-value structs instead.
	dtoResults := make(map[string]*api.RefSummary, len(results))
	for k, sum := range results {
		dtoResults[k] = toDTORefSummary(sum)
	}
	respondJSON(w, http.StatusOK, struct {
		Results map[string]*api.RefSummary `json:"results"`
	}{Results: dtoResults})
}

// toDTORefSummary converts a domain Summary to the wire DTO. Nil in,
// nil out -- the JSON encoder emits null for the entry.
func toDTORefSummary(s *refs.Summary) *api.RefSummary {
	if s == nil {
		return nil
	}

	dto := &api.RefSummary{
		Id:   openapi_types.UUID(s.ID),
		Type: api.RefSummaryType(s.Type),
	}

	switch s.Type {
	case refs.TypeStudyGuide:
		title := s.Title
		dto.Title = &title
		if s.Course != nil {
			dto.Course = &api.RefCourseInfo{
				Department: s.Course.Department,
				Number:     s.Course.Number,
			}
		}
		if s.QuizCount != nil {
			qc := *s.QuizCount
			dto.QuizCount = &qc
		}
		if s.IsRecommended != nil {
			rec := *s.IsRecommended
			dto.IsRecommended = &rec
		}
	case refs.TypeQuiz:
		title := s.Title
		dto.Title = &title
		if s.QuestionCount != nil {
			qc := *s.QuestionCount
			dto.QuestionCount = &qc
		}
		if s.Creator != nil {
			dto.Creator = &api.RefCreatorInfo{
				FirstName: s.Creator.FirstName,
				LastName:  s.Creator.LastName,
			}
		}
	case refs.TypeFile:
		name := s.Name
		dto.Name = &name
		if s.Size != nil {
			sz := *s.Size
			dto.Size = &sz
		}
		mime := s.MimeType
		dto.MimeType = &mime
		status := s.Status
		dto.Status = &status
	case refs.TypeCourse:
		title := s.Title
		dto.Title = &title
		dep := s.Department
		dto.Department = &dep
		num := s.Number
		dto.Number = &num
		if s.School != nil {
			dto.School = &api.RefSchoolInfo{
				Name:    s.School.Name,
				Acronym: s.School.Acronym,
			}
		}
	}

	return dto
}
