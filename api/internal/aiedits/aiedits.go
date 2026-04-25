// Package aiedits owns the persistence and business rules for the
// AI edit audit table (study_guide_edits, ASK-215). It records the
// (instruction, selection, replacement) tuple of every accepted or
// rejected AI rewrite so we have:
//
//   - cost / quota attribution beyond the raw ai_usage row
//   - eval signal (accepted/rejected ratio per feature, model)
//   - a future replay path if we want to re-evaluate old prompts
//
// The package owns nothing about prompt construction or model
// dispatch -- the handler builds the prompt, the ai package streams,
// and aiedits.Service is called once after the stream terminates.
package aiedits

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
)

// ErrNotFound is returned by Get / UpdateAcceptance when the row
// doesn't exist or doesn't belong to the requesting user. Mapped to
// HTTP 404 by the handler so callers can't probe other users' edit
// IDs by checking response codes.
var ErrNotFound = errors.New("aiedits: edit not found")

// Edit is the domain shape exposed to handlers + tests. Mirrors the
// db.StudyGuideEdit row but uses google/uuid.UUID + plain time.Time
// so callers don't pull pgtype into their tests.
type Edit struct {
	ID             uuid.UUID
	StudyGuideID   uuid.UUID
	UserID         uuid.UUID
	Instruction    string
	SelectionStart int32
	SelectionEnd   int32
	OriginalSpan   string
	Replacement    string
	Model          string
	InputTokens    int64
	OutputTokens   int64
	Accepted       *bool
	AcceptedAt     *time.Time
	CreatedAt      time.Time
}

// RecordEditParams is the input to RecordEdit. Built by the handler
// after the stream terminates -- the handler concatenates `delta`
// events into Replacement and reads InputTokens / OutputTokens off
// the final usage event.
type RecordEditParams struct {
	StudyGuideID   uuid.UUID
	UserID         uuid.UUID
	Instruction    string
	SelectionStart int32
	SelectionEnd   int32
	OriginalSpan   string
	Replacement    string
	Model          string
	InputTokens    int64
	OutputTokens   int64
}

// UpdateAcceptanceParams is the input to UpdateAcceptance. The PATCH
// handler converts the wire body into this shape; AcceptedAt is
// stamped here (not by the caller) so the time always matches when
// the row was actually marked.
type UpdateAcceptanceParams struct {
	ID           uuid.UUID
	StudyGuideID uuid.UUID
	UserID       uuid.UUID
	Accepted     bool
}

// Repository is the slice of db.Querier the service depends on.
// Defined here (where used) so tests can pass a fake.
type Repository interface {
	InsertEdit(ctx context.Context, params RecordEditParams) (Edit, error)
	UpdateAcceptance(ctx context.Context, params UpdateAcceptanceParams, at time.Time) (Edit, error)
	GetEdit(ctx context.Context, editID, studyGuideID uuid.UUID) (Edit, error)
}

// Service holds the AI-edit business rules. Construct via NewService.
// Safe for concurrent use.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService wires a Service over the given repository. Tests inject
// a deterministic clock via WithClock.
func NewService(repo Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option tunes a Service at construction.
type Option func(*Service)

// WithClock injects a deterministic time source -- only used by
// tests that assert AcceptedAt without flake.
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

// RecordEdit persists a row after a successful AI edit stream. The
// handler MUST call this on every successful Stream completion so
// the audit + eval signal stays complete.
func (s *Service) RecordEdit(ctx context.Context, params RecordEditParams) (Edit, error) {
	if err := params.validate(); err != nil {
		return Edit{}, err
	}
	row, err := s.repo.InsertEdit(ctx, params)
	if err != nil {
		return Edit{}, fmt.Errorf("aiedits: insert edit: %w", err)
	}
	return row, nil
}

// UpdateAcceptance records the user's accept/reject decision on an
// existing edit row. Stamps accepted_at to the service's clock so
// callers can't backdate by passing an old value.
//
// Returns ErrNotFound if the row doesn't exist OR if it exists but
// belongs to a different user / different guide -- callers can't
// distinguish "wrong row" from "no such row" by status code.
func (s *Service) UpdateAcceptance(ctx context.Context, params UpdateAcceptanceParams) (Edit, error) {
	if params.ID == uuid.Nil || params.StudyGuideID == uuid.Nil || params.UserID == uuid.Nil {
		return Edit{}, fmt.Errorf("aiedits: UpdateAcceptance requires non-nil IDs")
	}
	row, err := s.repo.UpdateAcceptance(ctx, params, s.now().UTC())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Edit{}, ErrNotFound
		}
		return Edit{}, fmt.Errorf("aiedits: update acceptance: %w", err)
	}
	return row, nil
}

func (p RecordEditParams) validate() error {
	if p.StudyGuideID == uuid.Nil {
		return fmt.Errorf("aiedits: StudyGuideID is required")
	}
	if p.UserID == uuid.Nil {
		return fmt.Errorf("aiedits: UserID is required")
	}
	if p.Instruction == "" {
		return fmt.Errorf("aiedits: Instruction is empty")
	}
	if p.SelectionStart < 0 || p.SelectionEnd < p.SelectionStart {
		return fmt.Errorf("aiedits: invalid selection range [%d, %d]", p.SelectionStart, p.SelectionEnd)
	}
	if p.OriginalSpan == "" {
		return fmt.Errorf("aiedits: OriginalSpan is empty")
	}
	if p.Replacement == "" {
		return fmt.Errorf("aiedits: Replacement is empty")
	}
	if p.Model == "" {
		return fmt.Errorf("aiedits: Model is required")
	}
	return nil
}

// rowToEdit narrows a sqlc db.StudyGuideEdit into our domain shape.
func rowToEdit(row db.StudyGuideEdit) Edit {
	out := Edit{
		ID:             uuid.UUID(row.ID.Bytes),
		StudyGuideID:   uuid.UUID(row.StudyGuideID.Bytes),
		UserID:         uuid.UUID(row.UserID.Bytes),
		Instruction:    row.Instruction,
		SelectionStart: row.SelectionStart,
		SelectionEnd:   row.SelectionEnd,
		OriginalSpan:   row.OriginalSpan,
		Replacement:    row.Replacement,
		Model:          row.Model,
		InputTokens:    row.InputTokens,
		OutputTokens:   row.OutputTokens,
	}
	if row.Accepted.Valid {
		v := row.Accepted.Bool
		out.Accepted = &v
	}
	if row.AcceptedAt.Valid {
		t := row.AcceptedAt.Time
		out.AcceptedAt = &t
	}
	if row.CreatedAt.Valid {
		out.CreatedAt = row.CreatedAt.Time
	}
	return out
}
