package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/dashboard"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

// DashboardService is the slice of the dashboard service surface
// this handler depends on. Defined here (where it is used) so the
// handler can be unit-tested with a mockery-generated mock without
// dragging in the full domain package.
type DashboardService interface {
	GetDashboard(ctx context.Context, p dashboard.GetDashboardParams) (dashboard.DashboardData, error)
}

// DashboardHandler serves GET /api/me/dashboard (ASK-155).
type DashboardHandler struct {
	service DashboardService
}

// NewDashboardHandler wires the handler over the given service.
func NewDashboardHandler(service DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// ListDashboard handles GET /me/dashboard. No query params.
func (h *DashboardHandler) ListDashboard(w http.ResponseWriter, r *http.Request) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	data, err := h.service.GetDashboard(r.Context(), dashboard.GetDashboardParams{
		ViewerID: viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListDashboard failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapDashboardResponse(data))
}

// mapDashboardResponse projects the domain DashboardData onto the
// wire envelope. All four sections are always non-nil; their inner
// arrays are always non-nil ([] not null).
func mapDashboardResponse(data dashboard.DashboardData) api.DashboardResponse {
	return api.DashboardResponse{
		Courses:     mapCoursesSection(data.Courses),
		StudyGuides: mapStudyGuidesSection(data.StudyGuides),
		Practice:    mapPracticeSection(data.Practice),
		Files:       mapFilesSection(data.Files),
	}
}

func mapCoursesSection(s dashboard.DashboardCoursesSection) api.DashboardCoursesSection {
	courses := make([]api.DashboardCourseSummary, 0, len(s.Courses))
	for _, c := range s.Courses {
		courses = append(courses, api.DashboardCourseSummary{
			Id:          c.ID,
			Department:  c.Department,
			Number:      c.Number,
			Title:       c.Title,
			Role:        api.DashboardCourseSummaryRole(c.Role),
			SectionTerm: c.SectionTerm,
		})
	}
	return api.DashboardCoursesSection{
		EnrolledCount: s.EnrolledCount,
		CurrentTerm:   s.CurrentTerm,
		Courses:       courses,
	}
}

func mapStudyGuidesSection(s dashboard.DashboardStudyGuidesSection) api.DashboardStudyGuidesSection {
	recent := make([]api.DashboardStudyGuideSummary, 0, len(s.Recent))
	for _, g := range s.Recent {
		recent = append(recent, api.DashboardStudyGuideSummary{
			Id:               g.ID,
			Title:            g.Title,
			CourseDepartment: g.CourseDepartment,
			CourseNumber:     g.CourseNumber,
			UpdatedAt:        g.UpdatedAt,
		})
	}
	return api.DashboardStudyGuidesSection{
		CreatedCount: s.CreatedCount,
		Recent:       recent,
	}
}

func mapPracticeSection(s dashboard.DashboardPracticeSection) api.DashboardPracticeSection {
	recent := make([]api.DashboardSessionSummary, 0, len(s.RecentSessions))
	for _, sess := range s.RecentSessions {
		recent = append(recent, api.DashboardSessionSummary{
			Id:              sess.ID,
			QuizTitle:       sess.QuizTitle,
			StudyGuideTitle: sess.StudyGuideTitle,
			ScorePercentage: sess.ScorePercentage,
			CompletedAt:     sess.CompletedAt,
		})
	}
	return api.DashboardPracticeSection{
		SessionsCompleted:      s.SessionsCompleted,
		TotalQuestionsAnswered: s.TotalQuestionsAnswered,
		OverallAccuracy:        s.OverallAccuracy,
		RecentSessions:         recent,
	}
}

func mapFilesSection(s dashboard.DashboardFilesSection) api.DashboardFilesSection {
	recent := make([]api.DashboardFileSummary, 0, len(s.Recent))
	for _, f := range s.Recent {
		recent = append(recent, api.DashboardFileSummary{
			Id:        f.ID,
			Name:      f.Name,
			MimeType:  f.MimeType,
			UpdatedAt: f.UpdatedAt,
		})
	}
	return api.DashboardFilesSection{
		TotalCount: s.TotalCount,
		TotalSize:  s.TotalSize,
		Recent:     recent,
	}
}
