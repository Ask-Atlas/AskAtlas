package handlers

import (
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
)

// CompositeHandler combines multiple handler structs into a single value
// that satisfies the generated api.ServerInterface.
type CompositeHandler struct {
	*FileHandler
	*GrantHandler
}

// NewCompositeHandler creates a handler that delegates to FileHandler and GrantHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler) *CompositeHandler {
	return &CompositeHandler{FileHandler: fh, GrantHandler: gh}
}

// ListSchools is a temporary stub so CompositeHandler satisfies the widened
// api.ServerInterface introduced by the new /schools operation. Removed in a
// follow-up commit when SchoolsHandler is embedded.
func (*CompositeHandler) ListSchools(w http.ResponseWriter, _ *http.Request, _ api.ListSchoolsParams) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
