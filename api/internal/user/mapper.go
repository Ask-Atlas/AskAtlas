package user

import (
	"encoding/json"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToUpsertClerkUserParams(payload UpsertUserPayload) (db.UpsertClerkUserParams, error) {
	var middleName pgtype.Text
	if payload.MiddleName != nil {
		middleName = pgtype.Text{String: *payload.MiddleName, Valid: true}
	}
	var metadata []byte
	if payload.Metadata != nil {
		b, err := json.Marshal(payload.Metadata)
		if err != nil {
			return db.UpsertClerkUserParams{}, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = b
	}
	return db.UpsertClerkUserParams{
		ClerkID:    payload.ClerkID,
		Email:      payload.Email,
		FirstName:  payload.FirstName,
		LastName:   payload.LastName,
		MiddleName: middleName,
		Metadata:   metadata,
	}, nil
}

func ToUser(dbUser db.User) (User, error) {
	var middleName *string
	if dbUser.MiddleName.Valid {
		middleName = &dbUser.MiddleName.String
	}
	var metadata map[string]interface{}
	if dbUser.Metadata != nil {
		if err := json.Unmarshal(dbUser.Metadata, &metadata); err != nil {
			return User{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	id, err := utils.PgxToGoogleUUID(dbUser.ID)
	if err != nil {
		return User{}, fmt.Errorf("failed to convert pgx uuid to google uuid: %w", err)
	}

	return User{
		ID:         id,
		ClerkID:    dbUser.ClerkID,
		Email:      dbUser.Email,
		FirstName:  dbUser.FirstName,
		LastName:   dbUser.LastName,
		MiddleName: middleName,
		Metadata:   metadata,
	}, nil
}
