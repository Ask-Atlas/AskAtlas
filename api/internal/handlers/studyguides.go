package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
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
	GetStudyGuide(ctx context.Context, params studyguides.GetStudyGuideParams) (studyguides.StudyGuideDetail, error)
	CreateStudyGuide(ctx context.Context, params studyguides.CreateStudyGuideParams) (studyguides.StudyGuideDetail, error)
	UpdateStudyGuide(ctx context.Context, params studyguides.UpdateStudyGuideParams) (studyguides.StudyGuideDetail, error)
	DeleteStudyGuide(ctx context.Context, params studyguides.DeleteStudyGuideParams) error
	CastVote(ctx context.Context, params studyguides.CastVoteParams) (studyguides.CastVoteResult, error)
	RemoveVote(ctx context.Context, params studyguides.RemoveVoteParams) error
	RecommendStudyGuide(ctx context.Context, params studyguides.RecommendStudyGuideParams) (studyguides.Recommendation, error)
	RemoveRecommendation(ctx context.Context, params studyguides.RemoveRecommendationParams) error
	AttachResource(ctx context.Context, params studyguides.AttachResourceParams) (studyguides.Resource, error)
	DetachResource(ctx context.Context, params studyguides.DetachResourceParams) error
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

// GetStudyGuide handles GET /study-guides/{study_guide_id}. Pure read
// per the design decision to keep GET idempotent + safe (HTTP
// semantics); view tracking is intentionally absent and will ship as
// a separate POST endpoint in a future ticket.
func (h *StudyGuideHandler) GetStudyGuide(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	detail, err := h.service.GetStudyGuide(r.Context(), studyguides.GetStudyGuideParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("GetStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapStudyGuideDetailResponse(detail))
}

// CreateStudyGuide handles POST /courses/{course_id}/study-guides.
// The body is decoded into the openapi-generated request type; the
// service layer applies tag normalization (trim + lowercase + dedupe)
// and runs the AssertCourseExists preflight so a missing course
// surfaces as 404 instead of an FK-violation 500. The creator id is
// always taken from the JWT -- the openapi schema explicitly forbids
// accepting one in the request body to avoid privilege-attribution
// forging.
func (h *StudyGuideHandler) CreateStudyGuide(w http.ResponseWriter, r *http.Request, courseId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CreateStudyGuideJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := studyguides.CreateStudyGuideParams{
		CourseID:    uuid.UUID(courseId),
		CreatorID:   viewerID,
		Title:       body.Title,
		Description: body.Description,
		Content:     body.Content,
	}
	if body.Tags != nil {
		params.Tags = append([]string(nil), *body.Tags...)
	}

	detail, err := h.service.CreateStudyGuide(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapStudyGuideDetailResponse(detail))
}

// UpdateStudyGuide handles PATCH /study-guides/{study_guide_id}.
// Decodes the partial-update body into pointer-typed params (so
// 'absent' is distinct from 'empty'), pulls viewer id from JWT,
// delegates to service.UpdateStudyGuide which gates on creator-only +
// returns the freshly re-hydrated detail. 200 on success; 400/403/404
// flow through ToHTTPError unchanged.
func (h *StudyGuideHandler) UpdateStudyGuide(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.UpdateStudyGuideJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
		Title:        body.Title,
		Description:  body.Description,
		Content:      body.Content,
	}
	if body.Tags != nil {
		// Snapshot the slice so the service can normalize it freely
		// without aliasing the decoded body's backing array.
		copied := append([]string(nil), (*body.Tags)...)
		params.Tags = &copied
	}

	detail, err := h.service.UpdateStudyGuide(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("UpdateStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapStudyGuideDetailResponse(detail))
}

// DeleteStudyGuide handles DELETE /study-guides/{study_guide_id}.
// Creator-only -- the service runs the locked SELECT + creator check
// + soft-delete + child-quiz cascade in a single transaction. 404
// covers both 'never existed' and 'already deleted' (idempotent
// semantics); 403 covers viewer-is-not-creator.
func (h *StudyGuideHandler) DeleteStudyGuide(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.DeleteStudyGuide(r.Context(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("DeleteStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CastStudyGuideVote handles POST /study-guides/{id}/votes.
// Decodes the body, validates that `vote` is one of "up" | "down" at
// the openapi-validator wrapper, then delegates to the service which
// upserts and returns the post-mutation vote_score so the UI can
// patch its local state without a follow-up GET.
func (h *StudyGuideHandler) CastStudyGuideVote(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CastStudyGuideVoteJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	result, err := h.service.CastVote(r.Context(), studyguides.CastVoteParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
		Vote:         studyguides.GuideVote(body.Vote),
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CastStudyGuideVote failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, api.CastVoteResponse{
		Vote:      api.CastVoteResponseVote(result.Vote),
		VoteScore: result.VoteScore,
	})
}

// RemoveStudyGuideVote handles DELETE /study-guides/{id}/votes.
// 404 covers both "guide missing/deleted" and "no existing vote" --
// the service already collapses both cases to apperrors.ErrNotFound.
func (h *StudyGuideHandler) RemoveStudyGuideVote(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.RemoveVote(r.Context(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RemoveStudyGuideVote failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RecommendStudyGuide handles POST /study-guides/{id}/recommendations.
// No request body. Service returns 403 / 404 / 409 directly via typed
// AppErrors; handler just maps + emits 201 with the recommendation
// payload on success.
func (h *StudyGuideHandler) RecommendStudyGuide(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	rec, err := h.service.RecommendStudyGuide(r.Context(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RecommendStudyGuide failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, api.RecommendationResponse{
		StudyGuideId:  openapi_types.UUID(rec.StudyGuideID),
		RecommendedBy: mapCreatorSummary(rec.Recommender),
		CreatedAt:     rec.CreatedAt,
	})
}

// RemoveStudyGuideRecommendation handles DELETE
// /study-guides/{id}/recommendations. Same role gate as POST. 204
// on success; 403/404 from the service flow through ToHTTPError.
func (h *StudyGuideHandler) RemoveStudyGuideRecommendation(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.RemoveRecommendation(r.Context(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RemoveStudyGuideRecommendation failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AttachResource handles POST /study-guides/{id}/resources.
// Decodes the body, takes attached_by from JWT, delegates to service.
// 201 on success with ResourceSummary; 409 on duplicate URL on guide;
// 404 on missing guide; 400 on validation; 500 on db errors.
func (h *StudyGuideHandler) AttachResource(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.AttachResourceJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := studyguides.AttachResourceParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		AttachedBy:   viewerID,
		Title:        body.Title,
		URL:          body.Url,
		Description:  body.Description,
	}
	if body.Type != nil {
		params.Type = studyguides.ResourceType(*body.Type)
	}

	resource, err := h.service.AttachResource(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("AttachResource failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapResourceSummaryResponse(resource))
}

// DetachResource handles DELETE /study-guides/{id}/resources/{resource_id}.
// 204 on success; 403 when viewer is neither guide creator nor
// attacher; 404 when guide is missing/deleted or resource isn't
// attached to this guide.
func (h *StudyGuideHandler) DetachResource(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID, resourceId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.DetachResource(r.Context(), studyguides.DetachResourceParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ResourceID:   uuid.UUID(resourceId),
		ViewerID:     viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("DetachResource failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
		Tags:          utils.NonNilStrings(g.Tags),
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

// mapGuideCourseSummaryResponse projects the compact
// GuideCourseSummary domain type onto the wire shape.
func mapGuideCourseSummaryResponse(c studyguides.GuideCourseSummary) api.GuideCourseSummary {
	return api.GuideCourseSummary{
		Id:         openapi_types.UUID(c.ID),
		Department: c.Department,
		Number:     c.Number,
		Title:      c.Title,
	}
}

// mapQuizSummaryResponse projects a domain Quiz onto the wire shape.
func mapQuizSummaryResponse(q studyguides.Quiz) api.QuizSummary {
	return api.QuizSummary{
		Id:            openapi_types.UUID(q.ID),
		Title:         q.Title,
		QuestionCount: q.QuestionCount,
	}
}

// mapResourceSummaryResponse projects a domain Resource onto the wire
// shape.
func mapResourceSummaryResponse(r studyguides.Resource) api.ResourceSummary {
	return api.ResourceSummary{
		Id:          openapi_types.UUID(r.ID),
		Title:       r.Title,
		Url:         r.URL,
		Type:        api.ResourceSummaryType(r.Type),
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
	}
}

// mapStudyGuideFileSummaryResponse projects a domain GuideFile onto
// the wire shape.
func mapStudyGuideFileSummaryResponse(f studyguides.GuideFile) api.StudyGuideFileSummary {
	return api.StudyGuideFileSummary{
		Id:       openapi_types.UUID(f.ID),
		Name:     f.Name,
		MimeType: f.MimeType,
		Size:     f.Size,
	}
}

// mapStudyGuideDetailResponse projects the full StudyGuideDetail
// domain type onto the get-by-id wire response. Every nested array
// (recommended_by, quizzes, resources, files) is emitted non-nil so
// the JSON output is '[]' rather than null when empty. user_vote is
// emitted as explicit JSON null when the viewer has not voted.
func mapStudyGuideDetailResponse(d studyguides.StudyGuideDetail) api.StudyGuideDetailResponse {
	recs := make([]api.CreatorSummary, 0, len(d.RecommendedBy))
	for _, r := range d.RecommendedBy {
		recs = append(recs, mapCreatorSummary(r))
	}
	quizzes := make([]api.QuizSummary, 0, len(d.Quizzes))
	for _, q := range d.Quizzes {
		quizzes = append(quizzes, mapQuizSummaryResponse(q))
	}
	resources := make([]api.ResourceSummary, 0, len(d.Resources))
	for _, r := range d.Resources {
		resources = append(resources, mapResourceSummaryResponse(r))
	}
	files := make([]api.StudyGuideFileSummary, 0, len(d.Files))
	for _, f := range d.Files {
		files = append(files, mapStudyGuideFileSummaryResponse(f))
	}

	resp := api.StudyGuideDetailResponse{
		Id:            openapi_types.UUID(d.ID),
		Title:         d.Title,
		Description:   d.Description,
		Content:       d.Content,
		Tags:          utils.NonNilStrings(d.Tags),
		Creator:       mapCreatorSummary(d.Creator),
		Course:        mapGuideCourseSummaryResponse(d.Course),
		VoteScore:     d.VoteScore,
		ViewCount:     d.ViewCount,
		IsRecommended: d.IsRecommended,
		RecommendedBy: recs,
		Quizzes:       quizzes,
		Resources:     resources,
		Files:         files,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	if d.UserVote != nil {
		uv := api.StudyGuideDetailResponseUserVote(*d.UserVote)
		resp.UserVote = &uv
	}
	return resp
}
