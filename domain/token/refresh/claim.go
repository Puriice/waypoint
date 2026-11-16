package refresh

import (
	"uuid"

	"github.com/puriice/waypoint/domain/errors"
	"github.com/puriice/waypoint/domain/user"
)

type Claims struct {
	SessionID uuid.UUID     `json:"session_id"`
	FamilyID  uuid.UUID     `json:"family_id"`
	Key       user.Password `json:"k"`
}

func NewClaims(sessionID, familyID string, key user.Password) (Claims, error) {
	sessionUUID, err := uuid.Parse(sessionID)

	if err != nil {
		return Claims{}, errors.NewErrInvalidUUID(sessionID, "session")
	}

	familyUUID, err := uuid.Parse(familyID)

	if err != nil {
		return Claims{}, errors.NewErrInvalidUUID(familyID, "family refresh token")
	}

	return Claims{
		SessionID: sessionUUID,
		FamilyID:  familyUUID,
		Key:       key,
	}, nil
}
