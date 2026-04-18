package handlers

import (
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CompositeHandler combines multiple handler structs into a single value
// that satisfies the generated api.ServerInterface.
type CompositeHandler struct {
	*FileHandler
	*GrantHandler
	*SchoolsHandler
}

// NewCompositeHandler creates a handler that delegates to FileHandler,
// GrantHandler, and SchoolsHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler, sh *SchoolsHandler) *CompositeHandler {
	return &CompositeHandler{FileHandler: fh, GrantHandler: gh, SchoolsHandler: sh}
}

// ListCourses is a temporary stub so CompositeHandler satisfies the widened
// api.ServerInterface introduced by the new /courses operation. Removed in a
// follow-up commit when CoursesHandler is embedded.
func (*CompositeHandler) ListCourses(w http.ResponseWriter, _ *http.Request, _ api.ListCoursesParams) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// GetCourse is a temporary stub paired with ListCourses above; same reason,
// same removal commit.
func (*CompositeHandler) GetCourse(w http.ResponseWriter, _ *http.Request, _ openapi_types.UUID) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
