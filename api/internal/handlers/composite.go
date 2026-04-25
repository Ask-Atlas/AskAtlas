package handlers

// CompositeHandler combines multiple handler structs into a single value
// that satisfies the generated api.ServerInterface.
type CompositeHandler struct {
	*FileHandler
	*GrantHandler
	*SchoolsHandler
	*CoursesHandler
	*StudyGuideHandler
	*StudyGuideGrantHandler
	*QuizzesHandler
	*SessionsHandler
	*RecentsHandler
	*FavoritesHandler
	*DashboardHandler
	*RefsHandler
	*AIHandler
	*AIEditHandler
}

// NewCompositeHandler creates a handler that delegates to every
// feature-specific handler.
func NewCompositeHandler(
	fh *FileHandler,
	gh *GrantHandler,
	sh *SchoolsHandler,
	ch *CoursesHandler,
	sgh *StudyGuideHandler,
	sggh *StudyGuideGrantHandler,
	qh *QuizzesHandler,
	ssh *SessionsHandler,
	rh *RecentsHandler,
	favh *FavoritesHandler,
	dh *DashboardHandler,
	refsh *RefsHandler,
	aih *AIHandler,
	aieh *AIEditHandler,
) *CompositeHandler {
	return &CompositeHandler{
		FileHandler:            fh,
		GrantHandler:           gh,
		SchoolsHandler:         sh,
		CoursesHandler:         ch,
		StudyGuideHandler:      sgh,
		StudyGuideGrantHandler: sggh,
		QuizzesHandler:         qh,
		SessionsHandler:        ssh,
		RecentsHandler:         rh,
		FavoritesHandler:       favh,
		DashboardHandler:       dh,
		RefsHandler:            refsh,
		AIHandler:              aih,
		AIEditHandler:          aieh,
	}
}
