package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestStreamRequest_Validate(t *testing.T) {
	t.Parallel()

	validUser := uuid.New()
	validMessages := []Message{
		{Role: RoleUser, Blocks: []Block{{Text: "hi"}}},
	}

	tests := []struct {
		name    string
		req     StreamRequest
		wantErr string
	}{
		{
			name: "valid",
			req: StreamRequest{
				UserID:   validUser,
				Feature:  FeaturePing,
				Messages: validMessages,
			},
		},
		{
			name: "missing user id",
			req: StreamRequest{
				Feature:  FeaturePing,
				Messages: validMessages,
			},
			wantErr: "UserID is required",
		},
		{
			name: "missing feature",
			req: StreamRequest{
				UserID:   validUser,
				Messages: validMessages,
			},
			wantErr: "Feature is required",
		},
		{
			name: "empty messages",
			req: StreamRequest{
				UserID:  validUser,
				Feature: FeaturePing,
			},
			wantErr: "Messages must contain at least one turn",
		},
		{
			name: "invalid role",
			req: StreamRequest{
				UserID:  validUser,
				Feature: FeaturePing,
				Messages: []Message{
					{Role: "system", Blocks: []Block{{Text: "hi"}}},
				},
			},
			wantErr: `Role "system" invalid`,
		},
		{
			name: "empty blocks",
			req: StreamRequest{
				UserID:  validUser,
				Feature: FeaturePing,
				Messages: []Message{
					{Role: RoleUser, Blocks: nil},
				},
			},
			wantErr: "Blocks is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestBuildParams_AppliesDefaults(t *testing.T) {
	t.Parallel()

	req := StreamRequest{
		UserID:  uuid.New(),
		Feature: FeaturePing,
		Messages: []Message{
			{Role: RoleUser, Blocks: []Block{{Text: "hi"}}},
		},
	}
	got := buildParams(req)

	if got.Model != ModelDefault {
		t.Errorf("Model = %q, want %q", got.Model, ModelDefault)
	}
	if !got.MaxCompletionTokens.Valid() || got.MaxCompletionTokens.Value != MaxTokensDefault {
		t.Errorf("MaxCompletionTokens = %v, want %d", got.MaxCompletionTokens, MaxTokensDefault)
	}
	if !got.StreamOptions.IncludeUsage.Valid() || !got.StreamOptions.IncludeUsage.Value {
		t.Errorf("StreamOptions.IncludeUsage = %v, want true (needed to surface token counts on the final chunk)", got.StreamOptions.IncludeUsage)
	}
	// One user message, no system block (none provided).
	if len(got.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(got.Messages))
	}
}

func TestBuildParams_SystemAndMessageOrder(t *testing.T) {
	t.Parallel()

	req := StreamRequest{
		UserID:  uuid.New(),
		Feature: FeatureGroundedEdit,
		System: []Block{
			{Text: "long stable system prompt", CacheControl: true},
			{Text: "short fragment"},
		},
		Messages: []Message{
			{Role: RoleUser, Blocks: []Block{{Text: "question"}}},
		},
	}
	got := buildParams(req)

	if len(got.Messages) != 2 {
		t.Fatalf("Messages len = %d, want 2 (system + user)", len(got.Messages))
	}
	if got.Messages[0].OfSystem == nil {
		t.Fatalf("Messages[0] is not a system message: %+v", got.Messages[0])
	}
	if got.Messages[1].OfUser == nil {
		t.Fatalf("Messages[1] is not a user message: %+v", got.Messages[1])
	}
}

// recordingRecorder captures every UsageRecord fed to it so tests
// can assert correct field translation from the cost-log hook.
type recordingRecorder struct {
	calls []UsageRecord
	err   error
}

func (r *recordingRecorder) RecordUsage(_ context.Context, rec UsageRecord) error {
	r.calls = append(r.calls, rec)
	return r.err
}

func TestClient_RecordUsage_Wired(t *testing.T) {
	t.Parallel()

	rec := &recordingRecorder{}
	c := NewClient("fake-key", nil, WithRecorder(rec))

	req := StreamRequest{
		UserID:  uuid.New(),
		Feature: FeatureEdit,
		Model:   ModelDefault,
		Messages: []Message{
			{Role: RoleUser, Blocks: []Block{{Text: "hi"}}},
		},
	}
	usage := Usage{InputTokens: 12, OutputTokens: 7, CacheReadTokens: 3}
	c.recordUsage(req, "req_test_123", usage)

	if len(rec.calls) != 1 {
		t.Fatalf("RecordUsage call count = %d, want 1", len(rec.calls))
	}
	got := rec.calls[0]
	if got.UserID != req.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, req.UserID)
	}
	if got.Feature != FeatureEdit {
		t.Errorf("Feature = %q, want %q", got.Feature, FeatureEdit)
	}
	if got.Model != ModelDefault {
		t.Errorf("Model = %q, want %q", got.Model, ModelDefault)
	}
	if got.Usage != usage {
		t.Errorf("Usage = %+v, want %+v", got.Usage, usage)
	}
	if got.RequestID != "req_test_123" {
		t.Errorf("RequestID = %q, want req_test_123", got.RequestID)
	}
}

func TestClient_RecordUsage_NoRecorderIsNoop(t *testing.T) {
	t.Parallel()

	// No WithRecorder option -- recorder is nil. Should not panic.
	c := NewClient("fake-key", nil)
	c.recordUsage(StreamRequest{
		Feature: FeaturePing,
		UserID:  uuid.New(),
	}, "req_xyz", Usage{})
}

func TestClient_RecordUsage_RecorderError_DoesNotPanic(t *testing.T) {
	t.Parallel()

	// A failing recorder should not crash the cost-log hook --
	// dropped row is a log line, never a panic.
	rec := &recordingRecorder{err: errFakeDB}
	c := NewClient("fake-key", nil, WithRecorder(rec))
	c.recordUsage(StreamRequest{
		Feature: FeaturePing,
		UserID:  uuid.New(),
	}, "req_err", Usage{})

	if len(rec.calls) != 1 {
		t.Errorf("call count = %d, want 1 (recorder still invoked)", len(rec.calls))
	}
}

var errFakeDB = errFake("db offline")

type errFake string

func (e errFake) Error() string { return string(e) }
