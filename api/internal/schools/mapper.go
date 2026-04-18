package schools

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
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
		Domain:    utils.TextPtr(r.Domain),
		URL:       utils.TextPtr(r.Url),
		City:      utils.TextPtr(r.City),
		State:     utils.TextPtr(r.State),
		Country:   utils.TextPtr(r.Country),
		CreatedAt: r.CreatedAt.Time,
	}, nil
}
