package user

import (
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestToUpserClerkUserParams(t *testing.T) {
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
				Metadata:   nil,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToUpsertClerkUserParams(tt.payload)
			if (err != nil) != tt.wantError {
				t.Errorf("ToUpsertClerkUserParams() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if got.ClerkID != tt.want.ClerkID {
				t.Errorf("ToUpsertClerkUserParams() ClerkID = %v, want %v", got.ClerkID, tt.want.ClerkID)
			}
			if got.Email != tt.want.Email {
				t.Errorf("ToUpsertClerkUserParams() Email = %v, want %v", got.Email, tt.want.Email)
			}
			if got.FirstName != tt.want.FirstName {
				t.Errorf("ToUpsertClerkUserParams() FirstName = %v, want %v", got.FirstName, tt.want.FirstName)
			}
			if got.LastName != tt.want.LastName {
				t.Errorf("ToUpsertClerkUserParams() LastName = %v, want %v", got.LastName, tt.want.LastName)
			}
			if got.MiddleName != tt.want.MiddleName {
				t.Errorf("ToUpsertClerkUserParams() MiddleName = %v, want %v", got.MiddleName, tt.want.MiddleName)
			}
			gotMeta, ok1 := got.Metadata.([]byte)
			wantMeta, ok2 := tt.want.Metadata.([]byte)

			if !ok1 && got.Metadata != nil {
				t.Errorf("got.Metadata is not []byte")
			}
			if !ok2 && tt.want.Metadata != nil {
				t.Errorf("tt.want.Metadata is not []byte")
			}

			if string(gotMeta) != string(wantMeta) {
				t.Errorf("ToUpsertClerkUserParams() Metadata = %s, want %s", gotMeta, wantMeta)
			}
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
			if (err != nil) != tt.wantError {
				t.Errorf("ToUser() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if got.ID != tt.want.ID {
				t.Errorf("ToUser() ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.ClerkID != tt.want.ClerkID {
				t.Errorf("ToUser() ClerkID = %v, want %v", got.ClerkID, tt.want.ClerkID)
			}
			if got.Email != tt.want.Email {
				t.Errorf("ToUser() Email = %v, want %v", got.Email, tt.want.Email)
			}
			if got.FirstName != tt.want.FirstName {
				t.Errorf("ToUser() FirstName = %v, want %v", got.FirstName, tt.want.FirstName)
			}
			if got.LastName != tt.want.LastName {
				t.Errorf("ToUser() LastName = %v, want %v", got.LastName, tt.want.LastName)
			}
			if (got.MiddleName == nil) != (tt.want.MiddleName == nil) || (got.MiddleName != nil && *got.MiddleName != *tt.want.MiddleName) {
				t.Errorf("ToUser() MiddleName = %v, want %v", got.MiddleName, tt.want.MiddleName)
			}
			if len(got.Metadata) != len(tt.want.Metadata) {
				t.Errorf("ToUser() Metadata length = %v, want %v", len(got.Metadata), len(tt.want.Metadata))
			} else {
				for k, v := range tt.want.Metadata {
					if gotVal, ok := got.Metadata[k]; !ok || gotVal != v {
						t.Errorf("ToUser() Metadata[%v] = %v, want %v", k, gotVal, v)
					}
				}
			}
		})
	}
}