package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// StudyGuideService defines the application logic required by the
// StudyGuideHandler. Mirrors CourseService: small, defined at the
// consumer, and mocked via mockery for handler tests.
type StudyGuideService interface {
	ListStudyGuides(ctx context.Context, params studyguides.ListStudyGuidesParams) (studyguides.ListStudyGuidesResult, error)
	AssertCourseExists(ctx context.Context, courseID uuid.UUID) error
}

// StudyGuideHandler manages incoming HTTP requests for the study-guide
// surface. Embedded in CompositeHandler so a single instance satisfies
// the generated api.ServerInterface.
type StudyGuideHandler struct {
	service StudyGuideService
}

// NewStudyGuideHandler creates a new StudyGuideHandler backed by the
// given StudyGuideService.
func NewStudyGuideHandler(service StudyGuideService) *StudyGuideHandler {
	return &StudyGuideHandler{service: service}
}

// ListStudyGuides handles GET /courses/{course_id}/study-guides.
// Runs the AssertCourseExists preflight first so a missing course
// surfaces as a tailored 404 'Course not found' (rather than an empty
// 200 that would be indistinguishable from 'course exists but has no
// guides'). Malformed cursors are rejected at the handler with a 400
// before the service is reached.
func (h *StudyGuideHandler) ListStudyGuides(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID, params api.ListStudyGuidesParams) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.AssertCourseExists(r.Context(), uuid.UUID(courseId)); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("AssertCourseExists failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	svcParams := studyguides.ListStudyGuidesParams{
		CourseID: uuid.UUID(courseId),
		Q:        params.Q,
	}
	if params.Tag != nil {
		svcParams.Tags = append([]string(nil), *params.Tag...)
	}
	if params.SortBy != nil {
		svcParams.SortBy = studyguides.SortField(*params.SortBy)
	}
	if params.SortDir != nil {
		svcParams.SortDir = studyguides.SortDir(*params.SortDir)
	}
	if params.PageLimit != nil {
		svcParams.Limit = int32(*params.PageLimit)
	}
	if params.Cursor != nil {
		cur, err := studyguides.DecodeCursor(*params.Cursor)
		if err != nil {
			apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			}))
			return
		}
		svcParams.Cursor = &cur
	}

	result, err := h.service.ListStudyGuides(r.Context(), svcParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListStudyGuides failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListStudyGuidesResponse(result))
}

// mapCreatorSummary projects the compact Creator domain type onto the
// wire shape.
func mapCreatorSummary(c studyguides.Creator) api.CreatorSummary {
	return api.CreatorSummary{
		Id:        openapi_types.UUID(c.ID),
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}
}

// mapStudyGuideListItemResponse projects a single StudyGuide to its
// list-row wire shape. Excludes content (only on the get-by-id
// endpoint) to keep the list payload small. Privacy floor: the nested
// creator payload is id + first_name + last_name only.
func mapStudyGuideListItemResponse(g studyguides.StudyGuide) api.StudyGuideListItemResponse {
	return api.StudyGuideListItemResponse{
		Id:            openapi_types.UUID(g.ID),
		Title:         g.Title,
		Description:   g.Description,
		Tags:          append([]string(nil), g.Tags...),
		Creator:       mapCreatorSummary(g.Creator),
		CourseId:      openapi_types.UUID(g.CourseID),
		VoteScore:     g.VoteScore,
		ViewCount:     g.ViewCount,
		IsRecommended: g.IsRecommended,
		QuizCount:     g.QuizCount,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

// mapListStudyGuidesResponse projects the domain result onto the
// paginated wire envelope. study_guides is always non-nil so the JSON
// output is '[]' rather than null when the course has no guides.
func mapListStudyGuidesResponse(r studyguides.ListStudyGuidesResult) api.ListStudyGuidesResponse {
	out := make([]api.StudyGuideListItemResponse, 0, len(r.StudyGuides))
	for _, g := range r.StudyGuides {
		out = append(out, mapStudyGuideListItemResponse(g))
	}
	return api.ListStudyGuidesResponse{
		StudyGuides: out,
		HasMore:     r.HasMore,
		NextCursor:  r.NextCursor,
	}
}
