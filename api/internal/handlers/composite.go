package handlers

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
