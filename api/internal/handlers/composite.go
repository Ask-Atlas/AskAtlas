package handlers

import (
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CompositeHandler combines multiple handler structs into a single value
// that satisfies the generated api.ServerInterface.
type CompositeHandler struct {
	*FileHandler
	*GrantHandler
	*SchoolsHandler
	*CoursesHandler
	*StudyGuideHandler
}

// NewCompositeHandler creates a handler that delegates to FileHandler,
// GrantHandler, SchoolsHandler, CoursesHandler, and StudyGuideHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler, sh *SchoolsHandler, ch *CoursesHandler, sgh *StudyGuideHandler) *CompositeHandler {
	return &CompositeHandler{
		FileHandler:       fh,
		GrantHandler:      gh,
		SchoolsHandler:    sh,
		CoursesHandler:    ch,
		StudyGuideHandler: sgh,
	}
}

// CreateStudyGuide is a temporary stub satisfying the generated
// ServerInterface while the studyguides.Service.CreateStudyGuide
// method + the real StudyGuideHandler wiring land in follow-up
// commits. Removed in commit 5 when the embedded *StudyGuideHandler
// provides the real method.
func (h *CompositeHandler) CreateStudyGuide(w http.ResponseWriter, _ *http.Request, _ openapi_types.UUID) {
	apperrors.RespondWithError(w, &apperrors.AppError{
		Code:    http.StatusNotImplemented,
		Status:  "Not Implemented",
		Message: "Endpoint not yet implemented",
	})
}

// DeleteStudyGuide is a temporary stub satisfying the generated
// ServerInterface while the studyguides.Service.DeleteStudyGuide
// method + the real StudyGuideHandler wiring land in follow-up
// commits. Removed in commit 5 when the embedded *StudyGuideHandler
// provides the real method.
func (h *CompositeHandler) DeleteStudyGuide(w http.ResponseWriter, _ *http.Request, _ openapi_types.UUID) {
	apperrors.RespondWithError(w, &apperrors.AppError{
		Code:    http.StatusNotImplemented,
		Status:  "Not Implemented",
		Message: "Endpoint not yet implemented",
	})
}

