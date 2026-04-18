package handlers

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
