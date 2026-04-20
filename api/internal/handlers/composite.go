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
	*RecentsHandler
	*FavoritesHandler
}

// NewCompositeHandler creates a handler that delegates to FileHandler,
// GrantHandler, SchoolsHandler, CoursesHandler, StudyGuideHandler,
// QuizzesHandler, SessionsHandler, RecentsHandler, and FavoritesHandler.
func NewCompositeHandler(fh *FileHandler, gh *GrantHandler, sh *SchoolsHandler, ch *CoursesHandler, sgh *StudyGuideHandler, qh *QuizzesHandler, ssh *SessionsHandler, rh *RecentsHandler, favh *FavoritesHandler) *CompositeHandler {
	return &CompositeHandler{
		FileHandler:       fh,
		GrantHandler:      gh,
		SchoolsHandler:    sh,
		CoursesHandler:    ch,
		StudyGuideHandler: sgh,
		QuizzesHandler:    qh,
		SessionsHandler:   ssh,
		RecentsHandler:    rh,
		FavoritesHandler:  favh,
	}
}
