package user

import (
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToUpsertClerkUserParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		payload   UpsertUserPayload
		want      db.UpsertClerkUserParams
		wantError bool
	}{
		{
			name: "All Fields Present",
			payload: UpsertUserPayload{
				ClerkID:    "clerk_123",
				Email:      "test@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				MiddleName: utils.Ptr("middle"),
				Metadata:   map[string]interface{}{"key": "value"},
			},
			want: db.UpsertClerkUserParams{
				ClerkID:    "clerk_123",
				Email:      "test@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				MiddleName: pgtype.Text{String: "middle", Valid: true},
				Metadata:   []byte(`{"key":"value"}`),
			},
			wantError: false,
		},
		{
			name: "Optional Fields Missing",
			payload: UpsertUserPayload{
				ClerkID:   "clerk_456",
				Email:     "jane@example.com",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			want: db.UpsertClerkUserParams{
				ClerkID:    "clerk_456",
				Email:      "jane@example.com",
				FirstName:  "Jane",
				LastName:   "Doe",
				MiddleName: pgtype.Text{Valid: false},
				Metadata:   []byte(nil),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToUpsertClerkUserParams(tt.payload)
			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.ClerkID, got.ClerkID)
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.FirstName, got.FirstName)
			assert.Equal(t, tt.want.LastName, got.LastName)
			assert.Equal(t, tt.want.MiddleName, got.MiddleName)

			// Handle metadata comparison specifically if needed, or rely on EqualValues if types match enough
			// For []byte content, Equal is usually fine if the content is deterministic
			assert.Equal(t, tt.want.Metadata, got.Metadata)
		})
	}
}

func TestToUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dbUser    db.User
		want      User
		wantError bool
	}{
		{
			name: "All Fields Present",
			dbUser: db.User{
				ID:         pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				ClerkID:    "clerk_123",
				Email:      "test@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				MiddleName: pgtype.Text{String: "middle", Valid: true},
				Metadata:   []byte(`{"key": "value"}`),
			},
			want: User{
				ID:         uuid.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				ClerkID:    "clerk_123",
				Email:      "test@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				MiddleName: utils.Ptr("middle"),
				Metadata:   map[string]interface{}{"key": "value"},
			},
			wantError: false,
		},
		{
			name: "Optional Fields Missing",
			dbUser: db.User{
				ID:         pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				ClerkID:    "clerk_456",
				Email:      "jane@example.com",
				FirstName:  "Jane",
				LastName:   "Doe",
				MiddleName: pgtype.Text{Valid: false},
				Metadata:   nil,
			},
			want: User{
				ID:         uuid.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				ClerkID:    "clerk_456",
				Email:      "jane@example.com",
				FirstName:  "Jane",
				LastName:   "Doe",
				MiddleName: nil,
				Metadata:   nil,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToUser(tt.dbUser)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.ClerkID, got.ClerkID)
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.FirstName, got.FirstName)
			assert.Equal(t, tt.want.LastName, got.LastName)
			assert.Equal(t, tt.want.MiddleName, got.MiddleName)
			assert.Equal(t, tt.want.Metadata, got.Metadata)
		})
	}
}
