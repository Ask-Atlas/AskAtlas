package clerk

import (
	"encoding/json"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/user"
)

func ParseWebhookEvent(event []byte) (Event, error) {
	var base BaseEvent
	if err := json.Unmarshal(event, &base); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}

	switch base.Type {
	case UserCreated:
		var userCreatedEvent UserCreatedEvent
		if err := json.Unmarshal(event, &userCreatedEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user created event: %w", err)
		}
		return userCreatedEvent, nil
	case UserUpdated:
		var userUpdateEvent UserUpdateEvent
		if err := json.Unmarshal(event, &userUpdateEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user update event: %w", err)
		}
		return userUpdateEvent, nil
	case UserDeleted:
		var userDeletedEvent UserDeletedEvent
		if err := json.Unmarshal(event, &userDeletedEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user deleted event: %w", err)
		}
		return userDeletedEvent, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", base.Type)
	}
}

const (
	MetadataKeyProfileImageURL = "profile_image_url"
	MetadataKeyImageURL        = "image_url"
	MetadataKeyHasImage        = "has_image"
	MetadataKeyClerkCreatedAt  = "clerk_created_at"
	MetadataKeyClerkUpdatedAt  = "clerk_updated_at"
	MetadataKeyLastSignInAt    = "last_sign_in_at"
	MetadataKeyLastActiveAt    = "last_active_at"
)

func ToUpsertUserPayload(clerkUser ClerkUser) (user.UpsertUserPayload, error) {
	emailAddress := clerkUser.GetPrimaryOrFirstEmailAddress()
	if emailAddress == nil {
		return user.UpsertUserPayload{}, fmt.Errorf("user has no primary email address")
	}

	var firstName, lastName string
	if clerkUser.FirstName != nil {
		firstName = *clerkUser.FirstName
	}
	if clerkUser.LastName != nil {
		lastName = *clerkUser.LastName
	}

	metadata := make(map[string]interface{})
	if clerkUser.ProfileImageURL != "" {
		metadata[MetadataKeyProfileImageURL] = clerkUser.ProfileImageURL
	}

	if clerkUser.ImageURL != "" {
		metadata[MetadataKeyImageURL] = clerkUser.ImageURL
	}

	metadata[MetadataKeyHasImage] = clerkUser.HasImage

	if clerkUser.PublicMetadata != nil {
		for k, v := range clerkUser.PublicMetadata {
			metadata[k] = v
		}
	}

	metadata[MetadataKeyClerkCreatedAt] = clerkUser.CreatedAt
	metadata[MetadataKeyClerkUpdatedAt] = clerkUser.UpdatedAt
	if clerkUser.LastSignInAt != nil {
		metadata[MetadataKeyLastSignInAt] = *clerkUser.LastSignInAt
	}
	if clerkUser.LastActiveAt != nil {
		metadata[MetadataKeyLastActiveAt] = *clerkUser.LastActiveAt
	}

	return user.UpsertUserPayload{
		ClerkID:    clerkUser.ID,
		Email:      emailAddress.EmailAddress,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: nil,
		Metadata:   metadata,
	}, nil
}
