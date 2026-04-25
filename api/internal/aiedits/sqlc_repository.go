package aiedits

import (
	"context"
	"errors"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SQLCRepository implements Repository over the sqlc-generated
// db.Queries. main.go wires one of these per process.
type SQLCRepository struct {
	q *db.Queries
}

// NewSQLCRepository wraps a db.Queries instance.
func NewSQLCRepository(q *db.Queries) *SQLCRepository {
	return &SQLCRepository{q: q}
}

func (r *SQLCRepository) InsertEdit(ctx context.Context, params RecordEditParams) (Edit, error) {
	row, err := r.q.InsertStudyGuideEdit(ctx, db.InsertStudyGuideEditParams{
		StudyGuideID:   utils.UUID(params.StudyGuideID),
		UserID:         utils.UUID(params.UserID),
		Instruction:    params.Instruction,
		SelectionStart: params.SelectionStart,
		SelectionEnd:   params.SelectionEnd,
		OriginalSpan:   params.OriginalSpan,
		Replacement:    params.Replacement,
		Model:          params.Model,
		InputTokens:    params.InputTokens,
		OutputTokens:   params.OutputTokens,
	})
	if err != nil {
		return Edit{}, err
	}
	return rowToEdit(row), nil
}

func (r *SQLCRepository) UpdateAcceptance(ctx context.Context, params UpdateAcceptanceParams, at time.Time) (Edit, error) {
	row, err := r.q.UpdateStudyGuideEditAcceptance(ctx, db.UpdateStudyGuideEditAcceptanceParams{
		ID:           utils.UUID(params.ID),
		StudyGuideID: utils.UUID(params.StudyGuideID),
		UserID:       utils.UUID(params.UserID),
		Accepted:     pgtype.Bool{Bool: params.Accepted, Valid: true},
		AcceptedAt:   pgtype.Timestamptz{Time: at, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Edit{}, ErrNotFound
		}
		return Edit{}, err
	}
	return rowToEdit(row), nil
}

func (r *SQLCRepository) GetEdit(ctx context.Context, editID, studyGuideID uuid.UUID) (Edit, error) {
	row, err := r.q.GetStudyGuideEdit(ctx, db.GetStudyGuideEditParams{
		ID:           utils.UUID(editID),
		StudyGuideID: utils.UUID(studyGuideID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Edit{}, ErrNotFound
		}
		return Edit{}, err
	}
	return rowToEdit(row), nil
}
