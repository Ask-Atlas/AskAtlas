package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
)

// fakeQuerier is a hand-rolled stub for QuotaQuerier. We avoid the
// mockery-generated mock here because the test surface is narrow
// (two methods, simple inputs) and a struct-based fake reads more
// clearly than `mockSvc.EXPECT().Method(...).Return(...)` chains.
type fakeQuerier struct {
	count    int64
	countErr error

	insertErr  error
	insertedAt int // monotonically incremented per InsertAIUsage call

	lastCountArg  db.CountAIUsageSinceParams
	lastInsertArg db.InsertAIUsageParams
}

func (f *fakeQuerier) CountAIUsageSince(_ context.Context, arg db.CountAIUsageSinceParams) (int64, error) {
	f.lastCountArg = arg
	return f.count, f.countErr
}

func (f *fakeQuerier) InsertAIUsage(_ context.Context, arg db.InsertAIUsageParams) (db.AiUsage, error) {
	f.lastInsertArg = arg
	f.insertedAt++
	if f.insertErr != nil {
		return db.AiUsage{}, f.insertErr
	}
	return db.AiUsage{}, nil
}

// frozenClock returns a function that always reports the given time.
// The quota service uses it to compute UTC midnight for the daily
// window; tests pin "now" so the math is deterministic.
func frozenClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestQuotaService_CheckAndReserve(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	now := time.Date(2026, 4, 25, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		quotas    Quotas
		feature   Feature
		count     int64
		wantOver  bool
		wantUsed  int64
		wantLimit int
	}{
		{
			name:    "under quota passes",
			quotas:  Quotas{FeatureEdit: 50},
			feature: FeatureEdit,
			count:   10,
		},
		{
			name:      "at quota rejects",
			quotas:    Quotas{FeatureEdit: 50},
			feature:   FeatureEdit,
			count:     50,
			wantOver:  true,
			wantUsed:  50,
			wantLimit: 50,
		},
		{
			name:      "over quota rejects",
			quotas:    Quotas{FeatureEdit: 50},
			feature:   FeatureEdit,
			count:     200,
			wantOver:  true,
			wantUsed:  200,
			wantLimit: 50,
		},
		{
			name:    "unconfigured feature is uncapped",
			quotas:  Quotas{FeatureEdit: 50},
			feature: FeatureQuiz,
			count:   1_000_000,
		},
		{
			name:    "zero limit is uncapped",
			quotas:  Quotas{FeatureEdit: 0},
			feature: FeatureEdit,
			count:   1_000_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeQuerier{count: tt.count}
			svc := NewQuotaService(fake, tt.quotas, WithClock(frozenClock(now)))

			err := svc.CheckAndReserve(context.Background(), user, tt.feature)

			if tt.wantOver {
				qe, ok := IsQuotaExceeded(err)
				if !ok {
					t.Fatalf("CheckAndReserve = %v, want *QuotaExceededError", err)
				}
				if qe.Used != tt.wantUsed || qe.Limit != tt.wantLimit {
					t.Errorf("got Used=%d Limit=%d, want Used=%d Limit=%d", qe.Used, qe.Limit, tt.wantUsed, tt.wantLimit)
				}
				wantReset := time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC)
				if !qe.ResetAt.Equal(wantReset) {
					t.Errorf("ResetAt = %v, want %v", qe.ResetAt, wantReset)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckAndReserve returned error: %v", err)
			}
		})
	}
}

func TestQuotaService_CheckAndReserve_DBError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("connection refused")
	svc := NewQuotaService(
		&fakeQuerier{countErr: wantErr},
		Quotas{FeatureEdit: 50},
	)
	err := svc.CheckAndReserve(context.Background(), uuid.New(), FeatureEdit)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CheckAndReserve err = %v, want wraps %v", err, wantErr)
	}
}

func TestQuotaService_RecordUsage(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	rec := UsageRecord{
		UserID:    user,
		Feature:   FeaturePing,
		Model:     ModelCheap,
		Usage:     Usage{InputTokens: 15, OutputTokens: 1, CacheReadTokens: 0},
		RequestID: "req_abc",
	}
	fake := &fakeQuerier{}
	svc := NewQuotaService(fake, nil)

	if err := svc.RecordUsage(context.Background(), rec); err != nil {
		t.Fatalf("RecordUsage returned error: %v", err)
	}
	if fake.insertedAt != 1 {
		t.Fatalf("insert call count = %d, want 1", fake.insertedAt)
	}
	got := fake.lastInsertArg
	if string(got.Feature) != string(FeaturePing) {
		t.Errorf("Feature = %q, want %q", got.Feature, FeaturePing)
	}
	if got.Model != string(ModelCheap) {
		t.Errorf("Model = %q, want %q", got.Model, ModelCheap)
	}
	if got.InputTokens != 15 || got.OutputTokens != 1 {
		t.Errorf("token counts = (%d, %d), want (15, 1)", got.InputTokens, got.OutputTokens)
	}
	if got.RequestID != "req_abc" {
		t.Errorf("RequestID = %q, want req_abc", got.RequestID)
	}
}

func TestQuotaService_RecordUsage_DBError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("constraint violation")
	svc := NewQuotaService(&fakeQuerier{insertErr: wantErr}, nil)
	err := svc.RecordUsage(context.Background(), UsageRecord{
		UserID:  uuid.New(),
		Feature: FeaturePing,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("RecordUsage err = %v, want wraps %v", err, wantErr)
	}
}

func TestUtcMidnight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "afternoon collapses to midnight UTC",
			in:   time.Date(2026, 4, 25, 14, 30, 45, 0, time.UTC),
			want: time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "non-UTC timezone normalizes",
			in:   time.Date(2026, 4, 25, 23, 30, 0, 0, time.FixedZone("est", -5*60*60)),
			want: time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utcMidnight(tt.in)
			if !got.Equal(tt.want) {
				t.Errorf("utcMidnight(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
