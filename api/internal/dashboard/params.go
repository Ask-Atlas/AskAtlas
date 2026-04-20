package dashboard

import "github.com/google/uuid"

const (
	// RecentGuidesLimit caps the number of guides returned in the
	// study_guides.recent array. Matches the spec's "Hardcoded
	// limits" section.
	RecentGuidesLimit int32 = 5
	// RecentSessionsLimit caps the number of sessions returned in
	// the practice.recent_sessions array.
	RecentSessionsLimit int32 = 5
	// RecentFilesLimit caps the number of files returned in the
	// files.recent array.
	RecentFilesLimit int32 = 5
	// MaxCourses caps the number of courses returned in the
	// courses.courses array. The dashboard is for the *current
	// term* and most users have <10 courses per term, so the cap
	// is a safety floor rather than a real product limit.
	MaxCourses int32 = 10
)

// GetDashboardParams is the input to Service.GetDashboard.
// ViewerID is resolved from the Clerk JWT in the handler before
// this struct is constructed; the service never sees the JWT
// directly. No filters or pagination on this endpoint -- the home
// page wants the same shape every render.
type GetDashboardParams struct {
	ViewerID uuid.UUID
}
