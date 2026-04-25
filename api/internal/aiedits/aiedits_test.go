package aiedits

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	insertErr  error
	updateErr  error
	getErr     error
	calls      int
	lastInsert RecordEditParams
	lastUpdate UpdateAcceptanceParams
	lastUpdAt  time.Time
}

func (f *fakeRepo) InsertEdit(_ context.Context, params RecordEditParams) (Edit, error) {
	f.calls++
	f.lastInsert = params
	if f.insertErr != nil {
		return Edit{}, f.insertErr
	}
	return Edit{
		ID:             uuid.New(),
		StudyGuideID:   params.StudyGuideID,
		UserID:         params.UserID,
		Instruction:    params.Instruction,
		SelectionStart: params.SelectionStart,
		SelectionEnd:   params.SelectionEnd,
		OriginalSpan:   params.OriginalSpan,
		Replacement:    params.Replacement,
		Model:          params.Model,
		InputTokens:    params.InputTokens,
		OutputTokens:   params.OutputTokens,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func (f *fakeRepo) UpdateAcceptance(_ context.Context, params UpdateAcceptanceParams, at time.Time) (Edit, error) {
	f.calls++
	f.lastUpdate = params
	f.lastUpdAt = at
	if f.updateErr != nil {
		return Edit{}, f.updateErr
	}
	acc := params.Accepted
	return Edit{
		ID:           params.ID,
		StudyGuideID: params.StudyGuideID,
		UserID:       params.UserID,
		Accepted:     &acc,
		AcceptedAt:   &at,
	}, nil
}

func (f *fakeRepo) GetEdit(_ context.Context, _, _ uuid.UUID) (Edit, error) {
	f.calls++
	if f.getErr != nil {
		return Edit{}, f.getErr
	}
	return Edit{}, nil
}

func validParams() RecordEditParams {
	return RecordEditParams{
		StudyGuideID:   uuid.New(),
		UserID:         uuid.New(),
		Instruction:    "make this clearer",
		SelectionStart: 100,
		SelectionEnd:   200,
		OriginalSpan:   "before",
		Replacement:    "after",
		Model:          "gpt-4.1",
		InputTokens:    50,
		OutputTokens:   25,
	}
}

func TestService_RecordEdit_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*RecordEditParams)
		wantErr string
	}{
		{name: "valid", mutate: func(*RecordEditParams) {}},
		{name: "missing study guide id", mutate: func(p *RecordEditParams) { p.StudyGuideID = uuid.Nil }, wantErr: "StudyGuideID is required"},
		{name: "missing user id", mutate: func(p *RecordEditParams) { p.UserID = uuid.Nil }, wantErr: "UserID is required"},
		{name: "empty instruction", mutate: func(p *RecordEditParams) { p.Instruction = "" }, wantErr: "Instruction is empty"},
		{name: "negative selection start", mutate: func(p *RecordEditParams) { p.SelectionStart = -1 }, wantErr: "invalid selection range"},
		{name: "inverted range", mutate: func(p *RecordEditParams) { p.SelectionEnd = p.SelectionStart - 1 }, wantErr: "invalid selection range"},
		{name: "empty original", mutate: func(p *RecordEditParams) { p.OriginalSpan = "" }, wantErr: "OriginalSpan is empty"},
		{name: "empty replacement", mutate: func(p *RecordEditParams) { p.Replacement = "" }, wantErr: "Replacement is empty"},
		{name: "missing model", mutate: func(p *RecordEditParams) { p.Model = "" }, wantErr: "Model is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			s := NewService(repo)
			p := validParams()
			tt.mutate(&p)
			_, err := s.RecordEdit(context.Background(), p)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RecordEdit returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want contains %q", err, tt.wantErr)
			}
		})
	}
}

func TestService_RecordEdit_PassesThroughToRepo(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	s := NewService(repo)
	p := validParams()

	got, err := s.RecordEdit(context.Background(), p)
	if err != nil {
		t.Fatalf("RecordEdit error: %v", err)
	}
	if repo.calls != 1 {
		t.Errorf("repo calls = %d, want 1", repo.calls)
	}
	if repo.lastInsert.Instruction != p.Instruction {
		t.Errorf("instruction not threaded: got %q, want %q", repo.lastInsert.Instruction, p.Instruction)
	}
	if got.Replacement != p.Replacement {
		t.Errorf("returned Edit.Replacement = %q, want %q", got.Replacement, p.Replacement)
	}
}

func TestService_UpdateAcceptance_StampsClock(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 18, 30, 0, 0, time.UTC)
	repo := &fakeRepo{}
	s := NewService(repo, WithClock(func() time.Time { return now }))

	editID, sgID, userID := uuid.New(), uuid.New(), uuid.New()
	got, err := s.UpdateAcceptance(context.Background(), UpdateAcceptanceParams{
		ID:           editID,
		StudyGuideID: sgID,
		UserID:       userID,
		Accepted:     true,
	})
	if err != nil {
		t.Fatalf("UpdateAcceptance error: %v", err)
	}
	if !repo.lastUpdAt.Equal(now) {
		t.Errorf("stamped at = %v, want %v", repo.lastUpdAt, now)
	}
	if got.AcceptedAt == nil || !got.AcceptedAt.Equal(now) {
		t.Errorf("returned AcceptedAt = %v, want %v", got.AcceptedAt, now)
	}
}

func TestService_UpdateAcceptance_NotFoundPropagates(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{updateErr: ErrNotFound}
	s := NewService(repo)
	_, err := s.UpdateAcceptance(context.Background(), UpdateAcceptanceParams{
		ID: uuid.New(), StudyGuideID: uuid.New(), UserID: uuid.New(),
	})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestService_UpdateAcceptance_NilIDsRejected(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	s := NewService(repo)
	_, err := s.UpdateAcceptance(context.Background(), UpdateAcceptanceParams{})
	if err == nil || !strings.Contains(err.Error(), "non-nil") {
		t.Errorf("err = %v, want contains 'non-nil'", err)
	}
}
