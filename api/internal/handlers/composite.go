package handlers

// CompositeHandler combines multiple handler structs into a single value
// that satisfies the generated api.ServerInterface.
type CompositeHandler struct {
	*FileHandler
	*GrantHandler
	*SchoolsHandler
	*CoursesHandler
	*StudyGuideHandler
	*QuizzesHandler
	*SessionsHandler
}

// NewCompositeHandler creates a handler that delegates to FileHandler,
// GrantHandler, SchoolsHandler, CoursesHandler, StudyGuideHandler,
// QuizzesHandler, and SessionsHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler, sh *SchoolsHandler, ch *CoursesHandler, sgh *StudyGuideHandler, qh *QuizzesHandler, ssh *SessionsHandler) *CompositeHandler {
	return &CompositeHandler{
		FileHandler:       fh,
		GrantHandler:      gh,
		SchoolsHandler:    sh,
		CoursesHandler:    ch,
		StudyGuideHandler: sgh,
		QuizzesHandler:    qh,
		SessionsHandler:   ssh,
	}
}
