package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgxToGoogleUUID(u pgtype.UUID) (uuid.UUID, error) {
    if !u.Valid {
        return uuid.Nil, fmt.Errorf("uuid is NULL/invalid")
    }
    return uuid.FromBytes(u.Bytes[:])
}