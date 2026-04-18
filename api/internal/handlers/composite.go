package handlers

import (
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
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
}

// NewCompositeHandler creates a handler that delegates to FileHandler,
// GrantHandler, SchoolsHandler, and CoursesHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler, sh *SchoolsHandler, ch *CoursesHandler) *CompositeHandler {
	return &CompositeHandler{
		FileHandler:    fh,
		GrantHandler:   gh,
		SchoolsHandler: sh,
		CoursesHandler: ch,
	}
}

// ListStudyGuides is a temporary stub satisfying the generated
// ServerInterface while the studyguides package + StudyGuideHandler
// land in follow-up commits. Will be removed when *StudyGuideHandler
// is embedded in CompositeHandler in commit 4 -- a real method on the
// embedded struct will satisfy the interface, and the explicit method
// here would otherwise shadow it.
func (h *CompositeHandler) ListStudyGuides(w http.ResponseWriter, _ *http.Request, _ openapi_types.UUID, _ api.ListStudyGuidesParams) {
	apperrors.RespondWithError(w, &apperrors.AppError{
		Code:    http.StatusNotImplemented,
		Status:  "Not Implemented",
		Message: "Endpoint not yet implemented",
	})
}
