package clerk_test

import (
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWebhookEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		event     []byte
		want      clerk.Event
		wantError bool
	}{
		{
			name: "UserCreated Event",
			event: []byte(`{
                "type": "user.created",
                "data": {
                    "id": "user_123",
                    "first_name": "John",
                    "last_name": "Doe"
                }
            }`),
			want: clerk.UserCreatedEvent{
				BaseEvent: clerk.BaseEvent{
					Type: clerk.UserCreated,
				},
				Data: clerk.ClerkUser{
					ID:        "user_123",
					FirstName: utils.Ptr("John"),
					LastName:  utils.Ptr("Doe"),
				},
			},
			wantError: false,
		},
		{
			name: "UserUpdated Event",
			event: []byte(`{
                "type": "user.updated",
                "data": {
                    "id": "user_123",
                    "username": "jdoe"
                }
            }`),
			want: clerk.UserUpdateEvent{
				BaseEvent: clerk.BaseEvent{
					Type: clerk.UserUpdated,
				},
				Data: clerk.ClerkUser{
					ID:       "user_123",
					Username: utils.Ptr("jdoe"),
				},
			},
			wantError: false,
		},
		{
			name: "UserDeleted Event",
			event: []byte(`{
                "type": "user.deleted",
                "data": {
                    "id": "user_123",
                    "deleted": true
                }
            }`),
			want: clerk.UserDeletedEvent{
				BaseEvent: clerk.BaseEvent{
					Type: clerk.UserDeleted,
				},
				Data: clerk.DeletedUser{
					ID:      "user_123",
					Deleted: true,
				},
			},
			wantError: false,
		},
		{
			name:      "Unknown Event Type",
			event:     []byte(`{"type": "unknown.event"}`),
			want:      nil,
			wantError: true,
		},
		{
			name:      "Invalid JSON",
			event:     []byte(`{invalid}`),
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := clerk.ParseWebhookEvent(tt.event)
			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToUpserUserPayload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		clerkUser clerk.ClerkUser
		wantEmail string
		wantError bool
	}{
		{
			name: "Primary Email Present",
			clerkUser: clerk.ClerkUser{
				ID: "user_123",
				EmailAddresses: []clerk.EmailAddress{
					{ID: "email_1", EmailAddress: "other@example.com"},
					{ID: "email_2", EmailAddress: "primary@example.com"},
				},
				PrimaryEmailAddressID: utils.Ptr("email_2"),
				FirstName:             utils.Ptr("John"),
			},
			wantEmail: "primary@example.com",
			wantError: false,
		},
		{
			name: "No Primary Email, Take First",
			clerkUser: clerk.ClerkUser{
				ID: "user_123",
				EmailAddresses: []clerk.EmailAddress{
					{ID: "email_1", EmailAddress: "first@example.com"},
				},
				FirstName: utils.Ptr("John"),
			},
			wantEmail: "first@example.com",
			wantError: false,
		},
		{
			name: "No Emails",
			clerkUser: clerk.ClerkUser{
				ID:             "user_123",
				EmailAddresses: []clerk.EmailAddress{},
			},
			wantEmail: "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := clerk.ToUpsertUserPayload(tt.clerkUser)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEmail, got.Email)
			assert.Equal(t, tt.clerkUser.ID, got.ClerkID)
		})
	}
}
