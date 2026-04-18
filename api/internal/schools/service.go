package schools

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service.
type Repository interface {
	ListSchools(ctx context.Context, arg db.ListSchoolsParams) ([]db.School, error)
}

// Service is the business-logic layer for the schools feature.
type Service struct {
	repo Repository
}

// NewService creates a new Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListSchools returns a paginated, optionally-filtered list of schools.
// The HTTP boundary is the primary validator (openapi enforces 1..MaxPageLimit),
// but the service also clamps defensively so internal Go callers can't ask
// Postgres for an unbounded number of rows.
func (s *Service) ListSchools(ctx context.Context, p ListSchoolsParams) (ListSchoolsResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	// Treat empty / whitespace-only q as no search.
	var qArg pgtype.Text
	if p.Q != nil {
		trimmed := strings.TrimSpace(*p.Q)
		if trimmed != "" {
			qArg = pgtype.Text{String: escapeLikePattern(trimmed), Valid: true}
		}
	}

	var cursorName pgtype.Text
	var cursorID pgtype.UUID
	if p.Cursor != nil {
		cursorName = pgtype.Text{String: p.Cursor.Name, Valid: true}
		cursorID = utils.UUID(p.Cursor.ID)
	}

	// Fetch limit+1 so we can tell whether there is a next page without a
	// second round-trip.
	rows, err := s.repo.ListSchools(ctx, db.ListSchoolsParams{
		Q:          qArg,
		CursorName: cursorName,
		CursorID:   cursorID,
		PageLimit:  limit + 1,
	})
	if err != nil {
		return ListSchoolsResult{}, fmt.Errorf("ListSchools: %w", err)
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}

	out := make([]School, 0, len(rows))
	for _, r := range rows {
		sch, err := mapSchool(r)
		if err != nil {
			return ListSchoolsResult{}, fmt.Errorf("ListSchools: %w", err)
		}
		out = append(out, sch)
	}

	var nextCursor *string
	if hasMore {
		// hasMore implies len(out) == limit >= 1, so out[len-1] is safe.
		last := out[len(out)-1]
		token, err := EncodeCursor(Cursor{Name: last.Name, ID: last.ID})
		if err != nil {
			return ListSchoolsResult{}, fmt.Errorf("ListSchools: encode cursor: %w", err)
		}
		nextCursor = &token
	}

	return ListSchoolsResult{
		Schools:    out,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// escapeLikePattern escapes the SQL LIKE/ILIKE wildcard characters %, _, and \
// using a backslash escape, matching the ESCAPE '\' clause in schools.sql.
// Without this, a user-provided q like "50%_off" would be interpreted as a
// wildcard pattern instead of a literal substring match.
func escapeLikePattern(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(s)
}
