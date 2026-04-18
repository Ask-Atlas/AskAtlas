package handlers

import (
	"context"
	"encoding/json"
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
	GetStudyGuide(ctx context.Context, params studyguides.GetStudyGuideParams) (studyguides.StudyGuideDetail, error)
	CreateStudyGuide(ctx context.Context, params studyguides.CreateStudyGuideParams) (studyguides.StudyGuideDetail, error)
	DeleteStudyGuide(ctx context.Context, params studyguides.DeleteStudyGuideParams) error
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
		Tags:          append([]string(nil), d.Tags...),
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
