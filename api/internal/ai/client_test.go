package ai

import (
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
