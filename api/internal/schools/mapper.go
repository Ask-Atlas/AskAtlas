package schools

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// mapSchool converts a sqlc-generated db.School row into the domain School type.
func mapSchool(r db.School) (School, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return School{}, fmt.Errorf("mapSchool: id: %w", err)
	}
	return School{
		ID:        id,
		Name:      r.Name,
		Acronym:   r.Acronym,
		Domain:    textPtr(r.Domain),
		URL:       textPtr(r.Url),
		City:      textPtr(r.City),
		State:     textPtr(r.State),
		Country:   textPtr(r.Country),
		CreatedAt: r.CreatedAt.Time,
	}, nil
}

// textPtr returns a *string for a nullable pgtype.Text column. Returns nil
// when the column is SQL NULL.
func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}
